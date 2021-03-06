package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PromonLogicalis/asn1"
	"github.com/PromonLogicalis/snmp"
)

var logger *log.Logger

// Convert OID in string format to OID in uint slice format
func strToOID(str string) (oid asn1.Oid, err error) {
	str = strings.TrimPrefix(str, ".") // remove leading dot
	subStrings := strings.Split(str, ".")
	oid = make(asn1.Oid, len(subStrings))
	for i, componentStr := range subStrings {
		x, err := strconv.ParseUint(componentStr, 10, 32)
		if err != nil {
			return nil, err
		}
		oid[i] = uint(x)
	}
	return oid, nil
}

func strToAddr(str string) (addr snmp.IPAddress, err error) {
	for i, component := range strings.Split(str, ".") {
		x, err := strconv.Atoi(component)
		if err != nil {
			return addr, err
		}
		addr[i] = byte(x)
	}
	return addr, nil
}

func convertBitsetToOctetStr(bitset BitsetMap) string {
	var maxK uint
	// get highest key in the set
	for k := range bitset {
		if k > maxK {
			maxK = k
		}
	}

	numBytes := maxK/8 + 1
	byteArr := make([]byte, numBytes)
	for k := range bitset {
		bytePos := k / 8
		bitPos := 7 - k%8
		byteArr[bytePos] |= 1 << bitPos
	}
	return string(byteArr)
}

func convertOctetStrToBitset(str string) (bitset BitsetMap) {
	bitset = make(BitsetMap)
	bytes := []byte(str)
	var j uint
	for i, b := range bytes {
		for j = 0; j < 8; j++ {
			if (b & (1 << j)) > 0 {
				bitset[uint(i)*8+j] = true
			}
		}
	}
	return bitset
}

func convertBytesToOctetStr(values map[string]uint, fieldInfo FieldInfo) (str string, err error) {
	bytes := make([]byte, fieldInfo.totalSize)

	// sort offsets
	offsArr := make([]int, 0)
	for offset := range fieldInfo.fieldOffsets {
		offsArr = append(offsArr, int(offset))
	}
	sort.Ints(offsArr)

	for _, o := range offsArr {
		offset := uint(o)
		id := fieldInfo.fieldOffsets[uint(offset)]
		size := fieldInfo.fieldSizes[id]
		value := values[id]

		// copy over data from 1 field value into bytes
		for dstIndex := offset; dstIndex < (offset + size); dstIndex++ {
			// Q: How about endianness - byte ordering?
			bytes[dstIndex] = byte(value & 0xFF)
			value = value >> 8
		}
	}
	str = string(bytes[:fieldInfo.totalSize])
	return str, nil
}

func addOIDFunc(agent *snmp.Agent, interp *Interpreter, strOid string, snmpMode SnmpMode) {
	if len(strOid) == 0 {
		logger.Println("Empty oid")
		return
	}
	oid, err := strToOID(strOid)
	if err != nil {
		logger.Printf("Bad oid %v (%s) - should not happen\n", oid, strOid)
		return
	}

	// given OID store away the provided value
	writeFunc := func(oid asn1.Oid, value interface{}) error {
		val := new(Value)
		oidStr := oid.String()
		typ := interp.variables.typesFromOid[oidStr]
		switch typ.valueType {
		case ValueString:
			switch value.(type) {
			case string:
				val.stringVal = value.(string)
			default:
				return errors.New("Bad string type")
			}
		case ValueInteger:
			switch value.(type) {
			case int:
				val.intVal = value.(int)
			default:
				return errors.New("Bad int type")
			}
		case ValueCounter:
			// Apparently one is not allowed to set a counter
			return errors.New("Cannot set counter type")
		case ValueBytes:
			return errors.New("Not supporting set bytes yet")
		case ValueTimeticks:
			switch value.(type) {
			case snmp.TimeTicks:
				val.intVal = int(value.(snmp.TimeTicks))
			default:
				return errors.New("Bad time ticks type")
			}
		case ValueGuage:
			switch value.(type) {
			case snmp.Unsigned32:
				val.intVal = int(value.(snmp.Unsigned32))
			default:
				return errors.New("Bad guage type")
			}
		case ValueOid:
			switch value.(type) {
			case asn1.Oid:
				oid := value.(asn1.Oid)
				val.oidVal = oid.String()
			default:
				return errors.New("Bad OID type")
			}
		case ValueIpv4address:
			switch value.(type) {
			case snmp.IPAddress:
				addr := value.(snmp.IPAddress)
				val.addrVal = addr.String()
			default:
				return errors.New("Bad ip address type")
			}
		case ValueBitset:
			switch value.(type) {
			case string:
				str := value.(string)
				val.bitsetVal = convertOctetStrToBitset(str)
			default:
				return errors.New("Bad bitset type")
			}
		}

		//fmt.Printf("received value of %v for oid %s\n", val, oidStr)
		switch snmpMode {
		case SnmpModeReadWrite:
			// update variable data under locking
			interp.SetValueForIdOid(typ.id, oidStr, val)
		case SnmpModeReadWriteBlocked:
			// use a blocking channel to send data
			typ.externalValue <- val
		}

		return nil
	}

	// given OID return its value
	readFunc := func(oid asn1.Oid) (interface{}, error) {
		oidStr := oid.String()
		//fmt.Printf("callback: oid: %s\n", oidStr)
		//fmt.Printf("oid values: %v\n", interp.oid2Values)
		val, found := interp.GetValueForOid(oidStr)
		if !found {
			return nil, errors.New("Illegal Value")
		}
		switch val.valueType {
		case ValueInteger:
			return val.intVal, nil
		case ValueCounter:
			return snmp.Counter32(val.intVal), nil
		case ValueTimeticks:
			return snmp.TimeTicks(val.intVal), nil
		case ValueGuage:
			return snmp.Unsigned32(val.intVal), nil
		case ValueString:
			return val.stringVal, nil
		case ValueBitset:
			return convertBitsetToOctetStr(val.bitsetVal), nil
		case ValueBytes:
			typ := interp.variables.typesFromOid[oidStr]
			return convertBytesToOctetStr(val.bytesVal, typ.fieldInfo)
		case ValueOid:
			oid, err := strToOID(val.oidVal)
			if err != nil {
				return nil, err
			}
			return oid, nil
		case ValueIpv4address:
			addr, err := strToAddr(val.addrVal)
			if err != nil {
				return nil, err
			}
			return addr, nil
		case ValueNone:
			return nil, errors.New("Illegal Value")
		}
		return nil, errors.New("Illegal Value")
	}

	switch snmpMode {
	case SnmpModeRead:
		agent.AddRoManagedObject(oid, readFunc)
	case SnmpModeReadWrite, SnmpModeReadWriteBlocked:
		agent.AddRwManagedObject(oid, readFunc, writeFunc)
	}
}

func initSNMPServer(interp *Interpreter, portNum uint, readCommunity string, writeCommunity string) (agent *snmp.Agent, conn *net.UDPConn, err error) {
	agent = snmp.NewAgent()

	// Set the read-only and read-write communities
	agent.SetCommunities(readCommunity, writeCommunity)

	// Bind to an UDP port
	portStr := ":" + strconv.FormatUint(uint64(portNum), 10)
	addr, err := net.ResolveUDPAddr("udp", portStr)
	if err != nil {
		return nil, nil, err
	}
	conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return nil, nil, err
	}

	//fmt.Printf("oid2Values: %v\n", interp.oid2Values)
	for oidStr := range interp.oid2Values {
		addOIDFunc(agent, interp, oidStr, interp.variables.typesFromOid[oidStr].snmpMode)
	}

	return agent, conn, err
}

// Read from a channel about OID requests
func runSNMPServer(agent *snmp.Agent, conn *net.UDPConn, quit chan bool, wg *sync.WaitGroup) {
	const readTimeoutSecs = 5

	defer wg.Done()

	// Serve requests
	for {

		// stop if told to finish up
		select {
		case <-quit:
			return
		default:
			// Do other stuff
		}

		// read incoming PDU
		buffer := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(readTimeoutSecs * time.Second))
		n, source, err := conn.ReadFrom(buffer)
		if err != nil {
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
				// error but not a network error or a network error other than timeout
				// handle non-timeout error
				logger.Printf("Failed to read buffer: %s", err)
				os.Exit(1)
			}
			// timeout => test for quit or try read again
			continue
		}

		// process PDU
		buffer, err = agent.ProcessDatagram(buffer[:n])
		if err != nil {
			logger.Println(err)
			continue
		}

		// respond with a new PDU
		_, err = conn.WriteTo(buffer, source)
		if err != nil {
			logger.Printf("Failed to write buffer: %s", err)
			os.Exit(1)
		}
	}
}

// -V key1=val1 -V key2=val2 -V key3=val3

func (varInits *VariableInits) String() string {
	return fmt.Sprintf("varinits: %v\n", *varInits)
}

// Set value for the variable initializations
// -V key1=val1 => value = "key1=val1"
func (varInits *VariableInits) Set(value string) error {
	strList := strings.Split(value, "=")
	if len(strList) != 2 {
		return errors.New("Invalid variable init")
	}
	(*varInits)[strList[0]] = strList[1]
	return nil
}

var version string // to be overridden with ldflags

// snmprun -p 161 -c public -C private -V key='value'
func main() {
	var portNum uint           // -p 161
	var readCommunity string   // -c public
	var writeCommunity string  // -C private
	var versionFlag bool       // -v
	var varInits VariableInits // -V key1=val1 -V key2=val2
	varInits = make(map[string]string)

	flag.UintVar(&portNum, "p", 161, "port number for SNMP server")
	flag.StringVar(&readCommunity, "c", "public", "community name")
	flag.StringVar(&writeCommunity, "C", "private", "community name")
	flag.BoolVar(&versionFlag, "v", false, "print version number")
	flag.Var(&varInits, "V", "variable initializers")
	flag.Parse()

	if versionFlag {
		if version == "" {
			version = "devel"
		}
		fmt.Printf("snmprun version %s\n", version)
		os.Exit(0)
	}

	if len(flag.Args()) != 1 {
		fmt.Print("Missing filename to run\n")
		os.Exit(1)
	}

	filename := flag.Args()[0]

	f, err := os.OpenFile(filename+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	logger = log.New(f, "snmpsim", log.LstdFlags)

	inputBuf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Unable to read file %s: %s\n", filename, err)
		os.Exit(1)
	}

	l := lex(filename, string(inputBuf))

	parser := NewParser(l)
	program, err := parser.ParseProgram()
	if err != nil {
		fmt.Printf("Parsing error: %s\n", err)
		os.Exit(1)
	}

	interp := new(Interpreter)
	interp.Init(program, varInits)

	agent, conn, err := initSNMPServer(interp, portNum, readCommunity, writeCommunity)
	if err != nil {
		fmt.Printf("Failed to init snmp server: %s\n", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	quitServer := make(chan bool)
	// SNMP server running in background
	go runSNMPServer(agent, conn, quitServer, &wg)

	// now run program to set the OID values
	err = interp.InterpProgram(program)
	if err != nil {
		logger.Printf("Interpreting error: %s\n", err)
	}
	quitServer <- true

	wg.Wait()
}
