package main

import "fmt"
import "bufio"
import "os"
import "strings"

func main() {
	fmt.Printf("%s\n%s\n", "GoClue is a cloud disk console client.",
		"Type \"login\" to sign up or \"h\" to get more help:")
	// var guessColor string
	// const favColor = "blue"
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		cmdString, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

	runCommand(cmdString)
		// err = runCommand(cmdString)
		// if err != nil {
		// 	fmt.Fprintln(os.Stderr, err)
		// }



		// fmt.Println("Guess my favorite color:")
		// if _, err := fmt.Scanf("%s", &guessColor); err != nil {
		// 	fmt.Printf("%s\n", err)
		// 	return
		// }
		// if favColor == guessColor {
		// 	fmt.Printf("%q is my favorite color!", favColor)
		// 	return
		// }
		// fmt.Printf("Sorry, %q is not my favorite color. Guess again. \n", guessColor)
	}

}

func runCommand(commandStr string) {
	commandStr = strings.TrimSuffix(commandStr, "\n")
	arrCommandStr := strings.Fields(commandStr)
	fmt.Printf("arrCommandStr: %d \n", len(arrCommandStr))
	switch arrCommandStr[0] {
	case "q":
		os.Exit(0)
	case "login":
		println("this is login")
	case "mkdir":
		println("this is mkdir")
	case "cd":
		println("this is cd")
	case "..":
		println("this is ..")
	case "rm":
		println("this is rm")
	case "d":
		println("this is download")
	case "ls":
		println("this is ls")
	case "u":
		println("this is upload")
	case "h":
		println("this is help")
	default:
		println("Please check your input or type \"h\" get help")
	}
}
