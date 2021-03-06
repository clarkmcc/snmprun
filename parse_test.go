package main

import "fmt"

func ExampleParse1() {
	inputStr := `var
	   test: string
	   x: integer
	   y: integer
	   i: integer
	   doit: boolean
	   this-boy: boolean
	endvar
	
	run
	    this-boy = true
	    x = 3 + 1000000 + 2 +
	          4 + 6
	    x = 3
		test = "hello"
		    
		y = x + 3
		print "y =  " + strInt(x)
		
		if y > 3
			test = "super"
		elseif y < 3 
		    test = "wonder"
		else
		    test = "duper"
		endif
		
		print test 
		test = test + "r"
		print test
		
		loop times 10
		    i = i + 1
		endloop
		
		print "i = " + strInt(i)
	endrun
	`

	// test input
	// inputStr := `var
	//    x: boolean
	//    y: boolean
	//    z: boolean
	//    a: integer
	//    b: integer
	//    c: string
	// endvar

	// run
	//x = true
	//y = true | false
	//y = x & (b | false)
	//z = y & false | x & (y | true)
	//z = y & false | x & y | true
	//z = (x | y & z) | (true & false | x)
	//    if x & y
	// 	  z = y
	// 	  z = true
	//    elseif true
	//       z = x | y
	//    elseif x | true
	//       z = y
	//    else
	// 	  z = false
	//    endif
	// loop x & y
	//    x = true
	// endloop
	//x = y | (z | true) & x
	//a = 2 / 3
	//a = 2 * 3
	//a = 3 + (b / a + 2) * 6
	//c = "hello" + " " + "there"
	//c = strInt(2 + 3 / 5)
	//print("hi" + " " + "there")
	//endrun`

	l := lex("test", inputStr)
	parser := NewParser(l)
	program, err := parser.ParseProgram()
	if err != nil {
		fmt.Print(err)
	} else {
		PrintProgram(program, 0)
	}
	// Output:
	// Program
	//   Variables
	//     Types
	//       doit: Boolean
	//       i: Integer
	//       test: String
	//       this-boy: Boolean
	//       x: Integer
	//       y: Integer
	//   StatementList
	//     Statement (type code: 2)
	//       Assignment
	//       lhs var = this-boy
	//         Expression
	//           Boolean Expression
	//           Or Terms
	//             [0]: term
	//               And Factors
	//                 [0]: factor
	//                 Const factor: true
	//     Statement (type code: 2)
	//       Assignment
	//       lhs var = x
	//         Expression
	//           Integer Expression
	//           Plus Terms
	//             [0]: plus term
	//               Times Factors
	//                 [0]: factor
	//                 Const factor: 3
	//             [1]: plus term
	//               Times Factors
	//                 [0]: factor
	//                 Const factor: 1000000
	//             [2]: plus term
	//               Times Factors
	//                 [0]: factor
	//                 Const factor: 2
	//             [3]: plus term
	//               Times Factors
	//                 [0]: factor
	//                 Const factor: 4
	//             [4]: plus term
	//               Times Factors
	//                 [0]: factor
	//                 Const factor: 6
	//     Statement (type code: 2)
	//       Assignment
	//       lhs var = x
	//         Expression
	//           Integer Expression
	//           Plus Terms
	//             [0]: plus term
	//               Times Factors
	//                 [0]: factor
	//                 Const factor: 3
	//     Statement (type code: 2)
	//       Assignment
	//       lhs var = test
	//         Expression
	//           String Expression
	//           Add String Terms
	//             [0]: string term
	//             Literal: "hello"
	//     Statement (type code: 2)
	//       Assignment
	//       lhs var = y
	//         Expression
	//           Integer Expression
	//           Plus Terms
	//             [0]: plus term
	//               Times Factors
	//                 [0]: factor
	//                 Id factor: x
	//             [1]: plus term
	//               Times Factors
	//                 [0]: factor
	//                 Const factor: 3
	//     Statement (type code: 3)
	//       Print Statement
	//         String Expression
	//         Add String Terms
	//           [0]: string term
	//           Literal: "y =  "
	//           [1]: string term
	//           Stringify Int Expression
	//             Integer Expression
	//             Plus Terms
	//               [0]: plus term
	//                 Times Factors
	//                   [0]: factor
	//                   Id factor: x
	//     Statement (type code: 1)
	//       If Statement
	//       predicate
	//         Boolean Expression
	//         Or Terms
	//           [0]: term
	//             And Factors
	//               [0]: factor
	//               Integer comparison
	//               Greater than >
	//                 Integer Expression
	//                 Plus Terms
	//                   [0]: plus term
	//                     Times Factors
	//                       [0]: factor
	//                       Id factor: y
	//                 Integer Expression
	//                 Plus Terms
	//                   [0]: plus term
	//                     Times Factors
	//                       [0]: factor
	//                       Const factor: 3
	//       if stmts
	//         StatementList
	//           Statement (type code: 2)
	//             Assignment
	//             lhs var = test
	//               Expression
	//                 String Expression
	//                 Add String Terms
	//                   [0]: string term
	//                   Literal: "super"
	//         [0] elsif
	//         elsif expression
	//           Boolean Expression
	//           Or Terms
	//             [0]: term
	//               And Factors
	//                 [0]: factor
	//                 Integer comparison
	//                 Less than <
	//                   Integer Expression
	//                   Plus Terms
	//                     [0]: plus term
	//                       Times Factors
	//                         [0]: factor
	//                         Id factor: y
	//                   Integer Expression
	//                   Plus Terms
	//                     [0]: plus term
	//                       Times Factors
	//                         [0]: factor
	//                         Const factor: 3
	//         elsif stmts
	//           StatementList
	//             Statement (type code: 2)
	//               Assignment
	//               lhs var = test
	//                 Expression
	//                   String Expression
	//                   Add String Terms
	//                     [0]: string term
	//                     Literal: "wonder"
	//       else stmts
	//         StatementList
	//           Statement (type code: 2)
	//             Assignment
	//             lhs var = test
	//               Expression
	//                 String Expression
	//                 Add String Terms
	//                   [0]: string term
	//                   Literal: "duper"
	//     Statement (type code: 3)
	//       Print Statement
	//         String Expression
	//         Add String Terms
	//           [0]: string term
	//           Identifier: test
	//     Statement (type code: 2)
	//       Assignment
	//       lhs var = test
	//         Expression
	//           String Expression
	//           Add String Terms
	//             [0]: string term
	//             Identifier: test
	//             [1]: string term
	//             Literal: "r"
	//     Statement (type code: 3)
	//       Print Statement
	//         String Expression
	//         Add String Terms
	//           [0]: string term
	//           Identifier: test
	//     Statement (type code: 0)
	//       Loop Statement (times)
	//         Integer Expression
	//         Plus Terms
	//           [0]: plus term
	//             Times Factors
	//               [0]: factor
	//               Const factor: 10
	//         StatementList
	//           Statement (type code: 2)
	//             Assignment
	//             lhs var = i
	//               Expression
	//                 Integer Expression
	//                 Plus Terms
	//                   [0]: plus term
	//                     Times Factors
	//                       [0]: factor
	//                       Id factor: i
	//                   [1]: plus term
	//                     Times Factors
	//                       [0]: factor
	//                       Const factor: 1
	//     Statement (type code: 3)
	//       Print Statement
	//         String Expression
	//         Add String Terms
	//           [0]: string term
	//           Literal: "i = "
	//           [1]: string term
	//           Stringify Int Expression
	//             Integer Expression
	//             Plus Terms
	//               [0]: plus term
	//                 Times Factors
	//                   [0]: factor
	//                   Id factor: i
	//
}

func ExampleParse2() {
	inputStr := `
	var
		printer-state: .1.3.3.2.1.1.1 integer [1 = 'printing', 2 = 'idle', 3 = 'error', ]
		printer-error: .1.3.2 bitset [1 = 'nopaper', 2 = 'notoner',]
		x: integer [1 = 'happy', 2 = 'sad']
	endvar
	run
	    printer-state = 1
	    printer-error = [ 'notoner', 'nopaper' ]
	endrun
	`
	l := lex("test", inputStr)
	parser := NewParser(l)
	program, err := parser.ParseProgram()
	if err != nil {
		fmt.Print(err)
	} else {
		PrintProgram(program, 0)
	}
	// Output:
	// Program
	//   Variables
	//     Types
	//       printer-error: Bitset oid: .1.3.2
	//       printer-state: Integer oid: .1.3.3.2.1.1.1
	//       x: Integer
	//     Aliases
	//       error: 3
	//       happy: 1
	//       idle: 2
	//       nopaper: 1
	//       notoner: 2
	//       printing: 1
	//       sad: 2
	//   StatementList
	//     Statement (type code: 2)
	//       Assignment
	//       lhs var = printer-state
	//         Expression
	//           Integer Expression
	//           Plus Terms
	//             [0]: plus term
	//               Times Factors
	//                 [0]: factor
	//                 Const factor: 1
	//     Statement (type code: 2)
	//       Assignment
	//       lhs var = printer-error
	//         Expression
	//           Bitset Expression
	//           Add terms
	//             [0]: bitset term
	//               [0]: bitset pos expression
	//                 Integer Expression
	//                 Plus Terms
	//                   [0]: plus term
	//                     Times Factors
	//                       [0]: factor
	//                       Const factor: 2
	//               [1]: bitset pos expression
	//                 Integer Expression
	//                 Plus Terms
	//                   [0]: plus term
	//                     Times Factors
	//                       [0]: factor
	//                       Const factor: 1
}
