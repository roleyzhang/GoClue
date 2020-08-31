package main

import (
	"flag"

	"github.com/golang/glog"

	// "encoding/json"
	"fmt"
	"time"

	// "io/ioutil"
	// "log"
	// "net/http"
	"os"
	// "os/exec"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"

	// "golang.org/x/net/context"
	// "golang.org/x/oauth2"
	// "golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	// "google.golang.org/api/option"
	. "github.com/logrusorgru/aurora"
	"github.com/roleyzhang/GoClue/cmd"
	"github.com/roleyzhang/GoClue/utils"
	"github.com/theckman/yacspin"
	// "github.com/roleyzhang/GoClue/utils"
	"bufio"
	"reflect"
)

func main() {
	flag.Parse()
	defer glog.Flush()
	// glog.V(8).Info("Level 8 log")
	// glog.V(5).Info("Level 5 log")
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		// prompt.OptionLivePrefix(changeLivePrefix),
		prompt.OptionLivePrefix(cmd.Ps.SetDynamicPrefix),
		prompt.OptionTitle("GOCULE"),
	)
	p.Run()

	// }else{
	// 	msg := fmt.Sprint(Brown("Please input Credentials.json file path"))
	// 	reader := bufio.NewReader(os.Stdin)
	// 	for {
	// 		fmt.Print("> ")
	// 		cmdString, err := reader.ReadString('\n')
	// 		if err != nil {
	// 			fmt.Fprintln(os.Stderr, err)
	// 		}

	// 		runCommand(cmdString)
	// 		// err = runCommand(cmdString)
	// 		// if err != nil {
	// 		// 	fmt.Fprintln(os.Stderr, err)
	// 		// }

	// 		// fmt.Println("Guess my favorite color:")
	// 		// if _, err := fmt.Scanf("%s", &guessColor); err != nil {
	// 		// 	fmt.Printf("%s\n", err)
	// 		// 	return
	// 		// }
	// 		// if favColor == guessColor {
	// 		// 	fmt.Printf("%q is my favorite color!", favColor)
	// 		// 	return
	// 		// }
	// 		// fmt.Printf("Sorry, %q is not my favorite color. Guess again. \n", guessColor)
	// 	}
	// }

	// err := flag.Lookup("logtostderr").Value.Set("true")
	// if err != nil{
	// 	glog.Error("Console issue")
	// }
	// // flag.Lookup("log_dir").Value.Set("/path/to/log/dir")
	// err = flag.Lookup("v").Value.Set("10")
	// if err != nil{
	// 	glog.Error("Console issue")
	// }
	//-----------------------THE OLD ONE
	// fmt.Printf("%s\n%s\n", "GoClue is a cloud disk console client.",
	// 	"Type \"login\" to sign up or \"h\" to get more help:")
	// // var guessColor string
	// // const favColor = "blue"
	// reader := bufio.NewReader(os.Stdin)

	// for {
	// 	fmt.Print("> ")
	// 	cmdString, err := reader.ReadString('\n')
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, err)
	// 	}

	// 	runCommand(cmdString)
	// 	// err = runCommand(cmdString)
	// 	// if err != nil {
	// 	// 	fmt.Fprintln(os.Stderr, err)
	// 	// }

	// 	// fmt.Println("Guess my favorite color:")
	// 	// if _, err := fmt.Scanf("%s", &guessColor); err != nil {
	// 	// 	fmt.Printf("%s\n", err)
	// 	// 	return
	// 	// }
	// 	// if favColor == guessColor {
	// 	// 	fmt.Printf("%q is my favorite color!", favColor)
	// 	// 	return
	// 	// }
	// 	// fmt.Printf("Sorry, %q is not my favorite color. Guess again. \n", guessColor)
	// }

}

func executor(in string) {
	runCommand(in)
	h := prompt.NewHistory()
	h.Add(in)

	// if in == "" {
	// 	LivePrefixState.IsEnable = false
	// 	LivePrefixState.LivePrefix = in
	// 	return
	// }
	// LivePrefixState.LivePrefix = in + ">>> "
	// LivePrefixState.IsEnable = true
}

func completer(in prompt.Document) []prompt.Suggest {
	// cmdStr = strings.TrimSuffix(cmdStr, "\n")
	dirSug = cmd.DirSug
	// fileSug = cmd.FileSug
	pathSug = cmd.PathSug
	allSug = cmd.AllSug
	idfileSug = cmd.IdfileSug
	iddirSug = cmd.IddirSug
	idallSug = cmd.IdAllSug
	typesSug = cmd.TypesSug
	roleSug = cmd.RoleSug
	gmailSug = cmd.GmailSug
	domainSug = cmd.DomainSug
	commentSug = cmd.CommentSug
	cmtListSug = cmd.CmtListSug
	arrCommandStr := strings.Fields(in.TextBeforeCursor())

	// fmt.Println("Your input: ",len(arrCommandStr) ,in.TextBeforeCursor())
	s := []prompt.Suggest{}

	if len(arrCommandStr) >= 0 {
		s = []prompt.Suggest{
			{Text: "q", Description: "Quit"},
			{Text: "login", Description: "Login to your account of net drive"},
			{Text: "mkdir", Description: "Create directory"},
			{Text: "rm", Description: "Delete directory or file"},
			{Text: "rmd", Description: "Delete directory or file by id"},
			{Text: "tr", Description: "Trash directory or file"},
			{Text: "trd", Description: "Trash directory or file by id"},
			{Text: "cd", Description: "change directory"},
			{Text: "cdd", Description: "change directory by id"},
			{Text: "mv", Description: "move file or directory, use \">\" Separate source and target"},
			{Text: "d", Description: "Download files use full path as save path and use \">\" Separate source and target"},
			{Text: "dd", Description: "Download files or directory by id use full path as save path and use \">\" Separate source and target"},
			{Text: "ls", Description: "list contents "},
			{Text: "share", Description: "share file/folder"},
			{Text: "shared", Description: "share file/folder by ID"},
			{Text: "comment", Description: "comment file/folder"},
			{Text: "commentd", Description: "commnet file/folder by ID"},
			{Text: "u", Description: "Upload directory or file, use \"-r\" for upload directory"},
			{Text: "h", Description: "Print help"},
			{Text: "n", Description: "Next page"},
			{Text: "p", Description: "Previous page"},
		}
	}
	if len(arrCommandStr) >= 1 {
		switch arrCommandStr[0] {
		case "d":
			// if !utils.IsContain(*allSug, in.GetWordBeforeCursorWithSpace()) {
			// 	// fmt.Println("cause u : ", in.GetWordBeforeCursorWithSpace())
			// 	cmd.PathFileGenerate(in.GetWordBeforeCursorWithSpace(), "4")
			// }
			if allSug != nil {
				s = *allSug
			}
		case "dd":
			if idfileSug != nil {
				s = *idallSug
			}
			// if allSug != nil {
			// 	s = *allSug
			// }
		case "rm":
			// if !utils.IsContain(*allSug, in.GetWordBeforeCursorWithSpace()) {
			// 	// fmt.Println("cause u : ", in.GetWordBeforeCursorWithSpace())
			// 	cmd.PathFileGenerate(in.GetWordBeforeCursorWithSpace(), "4")
			// }
			if allSug != nil {
				s = *allSug
			}
		case "rmd":
			if idfileSug != nil {
				s = *idallSug
			}
		case "tr":
			if allSug != nil {
				s = *allSug
			}
		case "trd":
			if idallSug != nil {
				s = *idallSug
			}
		case "mv":
			if !utils.IsContain(*allSug, in.GetWordBeforeCursorWithSpace()) {
				// fmt.Println("cause u : ", in.GetWordBeforeCursorWithSpace())
				cmd.PathFileGenerate(in.GetWordBeforeCursorWithSpace(), "4")
			}
			if allSug != nil {
				s = *allSug
			}
			// fmt.Println("cause mv : ", in.Text)
		case "cd":
			if dirSug != nil {
				s = *dirSug
			}
		case "cdd":
			if iddirSug != nil {
				s = *iddirSug
			}
		case "share":
			if allSug != nil {
				s = *allSug
			}
		case "shared":
			if idallSug != nil {
				s = *idallSug
			}
		case "comment":
			if allSug != nil {
				s = *allSug
			}
		case "commentd":
			if idallSug != nil {
				s = *idallSug
			}
		}
	}
	if len(in.TextBeforeCursor()) >= 2 && len(arrCommandStr) > 0 {
		switch arrCommandStr[0] {
		case "ls":
			s = []prompt.Suggest{
				{Text: "-t", Description: " filter by file type"},
				{Text: "-n", Description: " list by name"},
				{Text: "-c", Description: " list all files which include text"},
				{Text: "-d", Description: " list all folder"},
				{Text: "-dir", Description: " list files in folder"},
				{Text: "-l", Description: " list linked folder"},
				{Text: "-s", Description: " list starred folder"},
				{Text: "-tr", Description: " list trashed"},
			}
		}
	}
	if len(arrCommandStr) >= 2 {
		switch arrCommandStr[0] {
		case "d":
			if !utils.IsContain(*pathSug, in.GetWordBeforeCursorWithSpace()) {
				// fmt.Println("cause u : ", in.GetWordBeforeCursorWithSpace())
				cmd.PathFileGenerate(in.GetWordBeforeCursorWithSpace(), "2")
			}
			if pathSug != nil {
				s = *pathSug
			}
		case "dd":
			if !utils.IsContain(*pathSug, in.GetWordBeforeCursorWithSpace()) {
				// fmt.Println("cause u : ", in.GetWordBeforeCursorWithSpace())
				cmd.PathFileGenerate(in.GetWordBeforeCursorWithSpace(), "2")
			}
			if pathSug != nil {
				s = *pathSug
			}
		case "u":
			if !utils.IsContain(*allSug, in.GetWordBeforeCursorWithSpace()) {
				// fmt.Println("cause u : ", in.GetWordBeforeCursorWithSpace())
				cmd.PathFileGenerate(in.GetWordBeforeCursorWithSpace(), "4")
			}
			if allSug != nil {
				s = *allSug
			}
		case "share", "shared":
			if arrCommandStr[1] != "" {
				// fmt.Println("typesSug: ", len(*typesSug))
				if typesSug != nil {
					s = *typesSug
				}
			}
		case "comment", "commentd":
			if arrCommandStr[1] != "" {
				if commentSug != nil {
					s = *commentSug
				}
			}
		}
		switch arrCommandStr[1] {
		case "-t", "--t":
			s = []prompt.Suggest{
				{Text: "application/vnd.google-apps.video", Description: " Video file"},
				{Text: "video/mp4", Description: " MP4"},
				{Text: "application/vnd.google-apps.audio", Description: " Audio"},
				{Text: "application/vnd.google-apps.photo", Description: " Photo"},
				{Text: "image/jpeg", Description: " JPEG"},
				{Text: "image/gif", Description: " GIF"},
				{Text: "application/vnd.google-apps.document", Description: " Google Docs"},
				{Text: "application/vnd.google-apps.spreadsheet", Description: " Google Sheets"},
				{Text: "application/vnd.google-apps.form", Description: " Google Forms"},
				{Text: "application/vnd.google-apps.drawing", Description: " Google Drawing"},
				{Text: "application/vnd.google-apps.presentation", Description: " Google Slides"},
				{Text: "application/vnd.google-apps.script", Description: " Google Apps Scripts"},
				{Text: "application/pdf", Description: " pdf file"},
				{Text: "application/msword", Description: " MS Word"},
				{Text: "application/vnd.ms-excel", Description: " MS EXCEL"},
				{Text: "text/html", Description: " HTML"},
				{Text: "text/plain", Description: " TXT"},
				{Text: "application/x-javascript", Description: " Javascript"},
				{Text: "application/x-httpd-php", Description: " PHP"},
				{Text: "text/css", Description: " CSS"},
				{Text: "application/vnd.google-apps.drive-sdk", Description: " 3rd party shortcut"},
				{Text: "application/vnd.google-apps.file", Description: " Google Drive file"},
				{Text: "application/vnd.google-apps.folder", Description: " Google Drive folder"},
				{Text: "application/vnd.google-apps.fusiontable", Description: " Google Fusion Tables"},
				{Text: "application/vnd.google-apps.map", Description: " Google My Maps"},
				{Text: "application/vnd.google-apps.shortcut", Description: " Shortcut"},
				{Text: "application/vnd.google-apps.site", Description: " Google Sites"},
				{Text: "application/vnd.google-apps.unknown", Description: " unknown file type"},
				{Text: "application/x-shockwave-flash", Description: " Flash"},
				{Text: "apt", Description: " list starred folder"},
			}
		case "-dir", "--dir":
			if dirSug != nil {
				s = *dirSug
			}
		// case "-r":
		// 	if dirSug != nil {
		// 		s = *dirSug
		// 	}
		// 	if strings.Contains(arrCommandStr[0], "d") {
		// 		s = *iddirSug
		// 	}
		}
	}

	if len(arrCommandStr) >= 3 {
		switch arrCommandStr[0] {
		case "share", "shared":
			if roleSug != nil {
				s = *roleSug
			}
		case "comment", "commentd":
			if arrCommandStr[2] == "-d" || arrCommandStr[2] == "-u" || arrCommandStr[2] == "-g" {
				if cmtListSug != nil {
					s = *cmtListSug
				}
			}
		case "u":
			if !utils.IsContain(*allSug, in.GetWordBeforeCursorWithSpace()) {
				// fmt.Println("cause u : ", in.GetWordBeforeCursorWithSpace())
				cmd.PathFileGenerate(in.GetWordBeforeCursorWithSpace(), "4")
			}
			if allSug != nil {
				s = *allSug
			}
		case "d":
			if !utils.IsContain(*allSug, in.GetWordBeforeCursorWithSpace()) {
				// fmt.Println("cause u : ", in.GetWordBeforeCursorWithSpace())
				cmd.PathFileGenerate(in.GetWordBeforeCursorWithSpace(), "2")
			}
			if allSug != nil {
				s = *allSug
			}
		case "dd":
			if !utils.IsContain(*allSug, in.GetWordBeforeCursorWithSpace()) {
				// fmt.Println("cause u : ", in.GetWordBeforeCursorWithSpace())
				cmd.PathFileGenerate(in.GetWordBeforeCursorWithSpace(), "2")
			}
			if allSug != nil {
				s = *allSug
			}
			// if allSug != nil {
			// 	s = *allSug
			// }
		}
	}
	if len(arrCommandStr) >= 4 {
		switch arrCommandStr[0] {
		case "share", "shared":
			// fmt.Println("typesSug: ", len(*gmailSug))
			if gmailSug != nil {
				s = *gmailSug
			}
		}
	}
	if len(arrCommandStr) >= 5 {
		switch arrCommandStr[0] {
		case "share", "shared":
			// fmt.Println("typesSug22: ", len(*domainSug))
			if domainSug != nil {
				s = *domainSug
			}
		}
	}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

// var LivePrefixState struct {
// 	LivePrefix string
// 	IsEnable   bool
// }

// func changeLivePrefix() (string, bool) {
// 	return LivePrefixState.LivePrefix, LivePrefixState.IsEnable
// }

//--------------------------------------------
type command struct {
	name  string
	param string
	tip   string
}

var allCommands []command
var pageToken string
var counter int
var page map[int]string

var dirSug *[]prompt.Suggest
// var fileSug *[]prompt.Suggest
var pathSug *[]prompt.Suggest
var allSug *[]prompt.Suggest
var idfileSug *[]prompt.Suggest
var iddirSug *[]prompt.Suggest
var idallSug *[]prompt.Suggest
var typesSug *[]prompt.Suggest
var roleSug *[]prompt.Suggest
var gmailSug *[]prompt.Suggest
var domainSug *[]prompt.Suggest
var commentSug *[]prompt.Suggest
var cmtListSug *[]prompt.Suggest

var colorRed string
var cfg *yacspin.Config
var credentialsPath string

var ii cmd.ItemInfo

func init() {
	utils.CheckCredentials(waitting, start)
	// 	start()
	// } else {
	// 	utils.Check(to, waitting, start)
	// }
}

func waitting() {
	msg := fmt.Sprint(Brown("Please input credentials.json file path"))
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s\n", msg)
	//-----------------------Checking credentials
	for {
		fmt.Print("> ")
		cmdString, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		credentialsPath = strings.TrimRight(cmdString, "\n")

		home, _ := os.UserHomeDir()
		to := fmt.Sprint(home, string(os.PathSeparator),
			".local", string(os.PathSeparator),
			"goclue", string(os.PathSeparator),
			"credentials.json")
		if utils.Movefile(credentialsPath, to) {
			// fmt.Println("Guess my favorite color:",credentialsPath)
			start()
			return
		} else {
			msg := fmt.Sprint(Red("There is no credentials.json file"))
			fmt.Printf("%s\n", msg)
		}
	}
}
func start() {
	//-----yacspin-----------------
	cfg = &yacspin.Config{
		Frequency: 300 * time.Millisecond,
		// CharSet:           yacspin.CharSets[31],
		Suffix:            "", //+target,
		SuffixAutoColon:   true,
		ColorAll:          true,
		Message:           "",
		StopCharacter:     "✓",
		StopFailMessage:   "",
		StopFailCharacter: "✗",
		StopFailColors:    []string{"fgRed"},
		StopColors:        []string{"fgGreen"},
	}
	spinner, err := yacspin.New(*cfg)
	if err != nil {
		glog.Error("Spin run error", err.Error())
	}
	if err := spinner.Frequency(100 * time.Millisecond); err != nil {
		glog.Error("Spin run error", err.Error())
	}
	// msg := fmt.Sprintf("   ")
	spinner.Suffix("")
	spinner.StopCharacter("")
	msgs := fmt.Sprintf("... %s ", Brown("Type \"ls\" to list files or \"h\" to get more help:"))
	spinner.StopMessage(msgs)
	err = spinner.CharSet(yacspin.CharSets[26])
	// handle the error
	if err != nil {
		glog.V(8).Info("Spin run error", err.Error())
	}

	if err := spinner.Start(); err != nil {
		glog.V(8).Info("Spin start error", err.Error())
		// glog.Errorf("Spin start error %v", err)
		// return err
	}
	//-----yacspin-----------------
	//------Checking Linux system
	if !utils.IsCommandAvailable("tree") {
		// fmt.Println("output===", "yes")
		glog.Fatalln(Red("Please install tree firstly, then restart this program"))
	}
	//------Checking Linux system end

	allCommands = []command{
		{"q", "", "Quit"},
		{"mkdir", "", "Create directory"},
		{"rm", "", "Delete directory or file"},
		{"rmd", "", "Delete directory or file by id"},
		{"tr", "", "Trash directory or file, use \"-r\" for delete directory"},
		{"trd", "", "Trash directory or file by id, use \"-r\" for delete directory"},
		{"cd", "", "change directory"},
		{"mv", "", "move file or directory, use \">\" Separate source and target"},
		{"d", "", "Download files use \"-r\" for download directory"},
		{"ls", "-t filter by file type \n" +
			"\t-n list by name \n" +
			"\t-dir list files of folder\n" +
			"\t-c list all file's include text\n" +
			"\t-d list all folder \n" +
			"\t-l list linked folder \n" +
			"\t-s list starred folder \n" +
			"\t-tr list trashed \n",
			"\tlist contents "},
		{"u", "", "Upload directory or file, use \"-r\" for upload directory"},
		{"h", "", "Print help"},
		{"n", "", "Next page"},
		{"p", "", "Previous page"},
	}

	colorRed = "\033[31m%s\n"
	page = make(map[int]string)
	// for prompt suggest
	cmd.PathGenerate("HOME", "5")
	cmd.PathFileGenerate("HOME", "3")

	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
	ii = cmd.Ii
	cmd.Ps.GetRoot(&ii)
}

// run the command which input by user
func runCommand(commandStr string) {
	commandStr = strings.TrimSuffix(commandStr, "\n")
	arrCommandStr := strings.Fields(commandStr)
	// fmt.Printf("arrCommandStr: %d \n", len(arrCommandStr))
	if len(arrCommandStr) > 0 {
		switch arrCommandStr[0] {
		case "q":
			gs := reflect.ValueOf(gmailSug)
			ds := reflect.ValueOf(domainSug)
			// glog.V(8).Info(gs.IsNil()," : ",ds.IsNil())
			if !gs.IsNil() && !ds.IsNil() {
				if len(*gmailSug) > 0 && len(*domainSug) > 0 {
					utils.SaveProperty("mail", gmailSug)
					utils.SaveProperty("domain", domainSug)
				}
			}
			os.Exit(0)
		case "lo":
			cmd.Lo()
			// ii.Lo(drive.DriveScope)

			// if len(arrCommandStr[1:]) >0 {
			// }else{
			// 	fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			// }
		case "mkdir":
			if len(arrCommandStr[1:]) > 0 {
				if _, err := ii.CreateDir(arrCommandStr[1]); err != nil {
					glog.Error("Can not create folder" + err.Error())
				}
				cmd.Ps.SetPrefix("")
			} else {
				fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			}
		case "cd":
			if len(arrCommandStr[1:]) > 0 {
				ii.GetNode(commandStr)
				cmd.Ps.SetPrefix("")
			} else {
				fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			}
		case "cdd":
			if len(arrCommandStr[1:]) > 0 {
				ii.GetNoded(arrCommandStr[1])
				cmd.Ps.SetPrefix("")
			} else {
				fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			}
		case "mv":
			if len(arrCommandStr[1:]) > 0 {
				if err := ii.Move(commandStr); err != nil {
					glog.Error("Can not move file" + err.Error())
				}
				cmd.Ps.SetPrefix("")
			} else {
				fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			}
		case "tr":
			if len(arrCommandStr[1:]) > 0 {
				// if arrCommandStr[1] == "-r" {
				// 	if err := ii.Trash(arrCommandStr[2], arrCommandStr[1]); err != nil {
				// 		glog.Error("Can not delete folder" + err.Error())
				// 	}
				// } else {
				for _, value := range arrCommandStr[1:] {
					if err := ii.Trash(value); err != nil {
						glog.Error("Can not delete file" + err.Error())
					}
				}
				cmd.Ps.SetPrefix("")
			} else {
				fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			}
		case "trd":
			if len(arrCommandStr[1:]) > 0 {
				// if arrCommandStr[1] == "-r" {
				// 	if err := ii.Trashd(arrCommandStr[2], arrCommandStr[1]); err != nil {
				// 		glog.Error("Can not delete folder" + err.Error())
				// 	}
				// } else {
				for _, value := range arrCommandStr[1:] {
					if err := ii.Trashd(value); err != nil {
						glog.Error("Can not delete file" + err.Error())
					}
				}
				cmd.Ps.SetPrefix("")
			} else {
				fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			}
		case "rm":
			if len(arrCommandStr[1:]) > 0 {
				for _, value := range arrCommandStr[1:] {
					if err := ii.Rm(value); err != nil {
						glog.Error("Can not delete file" + err.Error())
					}
				}
				cmd.Ps.SetPrefix("")
			} else {
				fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			}
		case "rmd":
			if len(arrCommandStr[1:]) > 0 {
				for _, value := range arrCommandStr[1:] {
					if err := ii.Rmd(value); err != nil {
						glog.Error("Can not delete file" + err.Error())
					}
				}
				cmd.Ps.SetPrefix("")
			} else {
				fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			}
		case "d":
			if len(arrCommandStr[1:]) > 0 {
				err := ii.Download(commandStr)
				if err != nil {
					glog.Errorf("Unable to download files: %v", err.Error())
				}
				cmd.Ps.SetPrefix("")
			} else {
				fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			}
		case "dd":
			if len(arrCommandStr[1:]) > 0 {
				err := cmd.Downloadd(arrCommandStr)
				if err != nil {
					glog.Errorf("Unable to download files: %v", err.Error())
				}
				cmd.Ps.SetPrefix("")
			} else {
				fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			}
		case "ls":
			list(arrCommandStr)
			cmd.Ps.SetPrefix("")
		case "u":
			if len(arrCommandStr[1:]) > 0 {
				for _, value := range arrCommandStr[1:] {
					ii.UpLod(value, drive.DriveScope)
				}
				cmd.Ps.SetPrefix("")
			} else {
				fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			}
		case "h":
			for _, cmd := range allCommands {
				fmt.Printf("%6s: %s %s \n", cmd.name, cmd.param, cmd.tip)
			}
			cmd.Ps.SetPrefix("")
		case "n":
			counter++
			if page[counter] == "" {
				page[counter] = pageToken
			}
			next(counter)
			cmd.Ps.SetPrefix("- Page " + strconv.Itoa(counter))
		case "p":
			if counter > 0 {
				counter--
			}
			pageToken = page[counter]
			previous(counter)
			cmd.Ps.SetPrefix("- Page " + strconv.Itoa(counter))
		case "share":
			var domain string
			if len(arrCommandStr) < 5 {
				alert := fmt.Sprint(Red("Command incomplete, please use \"h\" get help"))
				fmt.Println(alert)
				return
			} else if len(arrCommandStr) == 5 {
				domain = ""
			} else {
				domain = arrCommandStr[5]
			}
			ii.Share(arrCommandStr[1], arrCommandStr[2], arrCommandStr[3], arrCommandStr[4], domain, true)
			cmd.Ps.SetPrefix("")
		case "shared":
			var domain string
			if len(arrCommandStr) < 5 {
				alert := fmt.Sprint(Red("Command incomplete, please use \"h\" get help"))
				fmt.Println(alert)
				return
			} else if len(arrCommandStr) == 5 {
				domain = ""
			} else {
				domain = arrCommandStr[5]
			}
			ii.Share(arrCommandStr[1], arrCommandStr[2], arrCommandStr[3], arrCommandStr[4], domain, false)
			cmd.Ps.SetPrefix("")
		case "comment":
			if len(arrCommandStr[1:]) > 0 {
				if arrCommandStr[2] == "-l" || arrCommandStr[2] == "-g" {
					ii.Commnet(arrCommandStr[1], arrCommandStr[2], "", "", true)
				} else if arrCommandStr[2] == "-u" {
					ii.Commnet(arrCommandStr[1], arrCommandStr[2], arrCommandStr[3], arrCommandStr[4], true)
				} else {
					ii.Commnet(arrCommandStr[1], arrCommandStr[2], arrCommandStr[3], "", true)
				}
				cmd.Ps.SetPrefix("")
			} else {
				fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			}
		case "commentd":
			if len(arrCommandStr[1:]) > 0 {
				if arrCommandStr[2] == "-l" || arrCommandStr[2] == "-g" {
					ii.Commnet(arrCommandStr[1], arrCommandStr[2], "", "", false)
				} else if arrCommandStr[2] == "-u" {
					ii.Commnet(arrCommandStr[1], arrCommandStr[2], arrCommandStr[3], arrCommandStr[4], false)
				} else {
					ii.Commnet(arrCommandStr[1], arrCommandStr[2], arrCommandStr[3], "", false)
				}
				cmd.Ps.SetPrefix("")
			} else {
				fmt.Printf(string(colorRed), "Command incomplete, please use \"h\" get help")
			}
		default:
			fmt.Printf(string(colorRed), "Please check your input or type \"h\" get help")
			cmd.Ps.SetPrefix("")
		}

	}
}

//------------

// setprefix ...
// func setPrefix(msgs string, ii *cmd.ItemInfo) {
// 	// folderId := ii.path[len(ii.path)-1]
// 	// fmt.Println(ii.itemId)
// 	folderId := ii.ItemId
// 	if dirSug != nil {
// 		folderName := cmd.GetSugDec(dirSug, folderId)
// 		msg(folderName + msgs)
// 	}
// }

// // msg ...
// func msg(message string) {
// 	LivePrefixState.LivePrefix = message + ">>> "
// 	LivePrefixState.IsEnable = true
// }

// func startSrv(scope string) *drive.Service {

// 	b, err := ioutil.ReadFile("credentials.json")
// 	if err != nil {
// 		glog.Errorln("Unable to read client secret file: %v", err)
// 	}

// 	// If modifying these scopes, delete your previously saved token.json.
// 	// config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
// 	// config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/drive")
// 	config, err := google.ConfigFromJSON(b, scope)
// 	if err != nil {
// 		glog.Errorln("Unable to parse client secret file to config: %v", err)
// 	}
// 	client := getClient(config)
// 	// client.Get(url string)
// 	// srv, err := drive.New(client)
// 	ctx := context.Background()
// 	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
// 	if err != nil {
// 		glog.Errorln("Unable to retrieve Drive client: %v", err)
// 	}
// 	return srv
// }

// list files of current directory
func list(cmds []string) {

	glog.V(8).Infof("cmds %s", cmds)
	// parameter setting
	// -dir list files of folder
	// -a show all type of items
	// -c show all type of items which include string of input
	// -d show all folder
	// -l show linked folder
	// -s show started folder
	// -t use file type to filter result
	// -n show by name
	if len(cmds) >= 2 {
		switch cmds[1] {
		case "-dir", "--dir":
			if len(cmds) == 3 {
				counter = 0
				clearMap()
				userQuery("dir", cmds[2])
			}
		case "-d", "--d":
			counter = 0
			clearMap()
			userQuery("d", "")
		case "-l", "--l":
			counter = 0
			clearMap()
			userQuery("l", "")
		case "-s", "--s":
			// qString = "starred"
			counter = 0
			clearMap()
			userQuery("s", "")
		case "-t", "--t":
			if len(cmds) == 3 {
				counter = 0
				clearMap()
				userQuery("t", cmds[2])
			}
		case "-n", "--n":
			if len(cmds) == 3 {
				// glog.V(8).Infof("qString len %d qString%s\n",len(cmds), qString)
				counter = 0
				clearMap()
				userQuery("n", cmds[2])
			}
		case "-c", "--c":
			if len(cmds) == 3 {
				counter = 0
				clearMap()
				userQuery("c", cmds[2])
			}
		case "-tr", "--trash":
			counter = 0
			clearMap()
			userQuery("tr", "")
		case "-a", "--all":
			counter = 0
			clearMap()
			userQuery("default", "")
		default:
			counter = 0
			clearMap()
			userQuery("default", "")
		}
	} else {
		// qString = "'" + ii.ItemId + "' in parents and trashed=false"
		counter = 0
		clearMap()
		userQuery("dls", ii.ItemId)
	}
}

// clear page map for new query
// clearMap ...
func clearMap() {
	for k := range page {
		delete(page, k)
	}
}

// base query
// name ...
func userQuery(param, cmd string) *drive.FileList {
	r := ii.ShowResult(page, counter, param, cmd, drive.DriveScope)
	if r == nil {
		fmt.Printf(string(colorRed), "No Result return: ")
		return nil
	}
	pageToken = r.NextPageToken
	return r
}

// show next page
func next(counter int) {
	r := ii.ShowResult(page, counter, "next", "", drive.DriveScope)
	if r == nil {
		fmt.Printf(string(colorRed), "No Result return: ")
	} else {
		pageToken = r.NextPageToken
	}
}

// show previous page
func previous(counter int) {
	r := ii.ShowResult(page, counter, "previous", "", drive.DriveScope)
	if r == nil {
		fmt.Printf(string(colorRed), "No Result return: ")
	}
}
