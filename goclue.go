package main

import (
	// "bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"os/exec"

	"github.com/c-bata/go-prompt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"github.com/dustin/go-humanize"
	"io"
)

func main() {
	fmt.Printf("%s\n%s\n", "GoClue is a cloud disk console client.",
		"Type \"login\" to sign up or \"h\" to get more help:")
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("GOCULE"),
	)
	p.Run()
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
}

func completer(in prompt.Document) []prompt.Suggest {
	// cmdStr = strings.TrimSuffix(cmdStr, "\n")
	arrCommandStr := strings.Fields(in.TextBeforeCursor())
	// fmt.Println("Your input: " , arrCommandStr )
	s := []prompt.Suggest{
		{Text: "q", Description: "Quit"},
		{Text: "login", Description: "Login to your account of net drive"},
		{Text: "mkdir", Description: "Create directory"},
		{Text: "rm", Description: "Delete directory or file, use \"-r\" for delete directory"},
		{Text: "cd", Description: "change directory"},
		{Text: "move", Description: "move file or directory"},
		{Text: "d", Description: "Download files use \"-r\" for download directory"},
		{Text: "ls", Description: "list contents "},
		{Text: "u", Description: "Upload directory or file, use \"-r\" for upload directory"},
		{Text: "h", Description: "Print help"},
		{Text: "n", Description: "Next page"},
		{Text: "p", Description: "Previous page"},
	}

	if len(arrCommandStr) == 1 {
		switch arrCommandStr[0]{
		case "d":
			if fileSug != nil {
				s = *fileSug
			}
		case "rm":
			if fileSug != nil {
				s = *fileSug
			}
		case "mv":
			if fileSug != nil {
				s = *fileSug
			}
		}
	}
	if len(arrCommandStr) >= 2 {
		s = []prompt.Suggest{
			{Text: "-t", Description: " filter by file type"},
			{Text: "-n", Description: " list by name"},
			{Text: "-d", Description: " list all folder"},
			{Text: "-dir", Description: " list files of folder"},
			{Text: "-l", Description: " list linked folder"},
			{Text: "-s", Description: " list starred folder"},
		}
		switch arrCommandStr[0] {
		case "d":
			if pathSug != nil {
				s = *pathSug
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
				{Text: "appt", Description: " list starred folder"},
			}
		case "-dir", "--dir":
			if dirSug != nil {
				s = *dirSug
			}
		}
	}
	// else if len(arrCommandStr) == 3 {
	// 	s = []prompt.Suggest{
	// 		{Text: "q", Description: "Quit"},
	// 		{Text: "login", Description: "Login to your account of net drive"},
	// 		{Text: "mkdir", Description: "Create directory"},
	// 		{Text: "rm", Description: "Delete directory or file, use \"-r\" for delete directory"},
	// 		{Text: "cd", Description: "change directory"},
	// 		{Text: "..", Description: "Exit current directory"},
	// 		{Text: "d", Description: "Download files use \"-r\" for download directory"},
	// 		{Text: "ls", Description: "\tlist contents "},
	// 		{Text: "u", Description: "Upload directory or file, use \"-r\" for upload directory"},
	// 		{Text: "h", Description: "Print help"},
	// 		{Text: "n", Description: "Next page"},
	// 		{Text: "p", Description: "Previous page"},
	// 	}
	// }
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

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
var qString string
var dirSug *[]prompt.Suggest
var fileSug *[]prompt.Suggest
var pathSug *[]prompt.Suggest

// var service *drive.Service

func init() {
	// fmt.Println("This will get called on main initialization")
	// allCommands = make([]command, 0)
	allCommands = []command{
		{"q", "", "Quit"},
		{"login", "", "Login to your account of net drive"},
		{"mkdir", "", "Create directory"},
		{"rm", "", "Delete directory or file, use \"-r\" for delete directory"},
		{"cd", "", "change directory"},
		{"move", "", "move file or directory"},
		{"d", "", "Download files use \"-r\" for download directory"},
		{"ls", "-t filter by file type \n" +
			"\t-n list by name \n" +
			"\t-dir list files of folder\n" +
			"\t-d list all folder \n" +
			"\t-l list linked folder \n" +
			"\t-s list starred folder \n",
			"\tlist contents "},
		{"u", "", "Upload directory or file, use \"-r\" for upload directory"},
		{"h", "", "Print help"},
		{"n", "", "Next page"},
		{"p", "", "Previous page"},
	}

	page = make(map[int]string)
	// for prompt suggest
	pathGenerate()	
}

// run the command which input by user
func runCommand(commandStr string) {
	commandStr = strings.TrimSuffix(commandStr, "\n")
	arrCommandStr := strings.Fields(commandStr)
	// fmt.Printf("arrCommandStr: %d \n", len(arrCommandStr))
	if len(arrCommandStr) > 0 {
		switch arrCommandStr[0] {
		case "q":
			os.Exit(0)
		case "login":
			// service = startSrv()
			println("this is login")
		case "mkdir":
			println("this is mkdir")
		case "cd":
			println("this is cd")
		case "mv":
			move()
		case "rm":
			rm()
		case "d":
			err := download(arrCommandStr)
			if err != nil {
				log.Fatalf("Unable to download files: %v", err)
			}
		case "ls":
			list(arrCommandStr)
			// println("this is ls")
		case "u":
			println("this is upload")
		case "h":
			for _, cmd := range allCommands {
				fmt.Printf("%6s: %s %s \n", cmd.name, cmd.param, cmd.tip)
			}
		case "n":
			counter++
			// fmt.Printf("counter %d", counter)
			if page[counter] == "" {
				page[counter] = pageToken
			}
			next(counter)
			fmt.Printf("page %d", counter)
		case "p":
			if counter > 0 {
				counter--
			}
			// fmt.Printf("counter %d", counter)
			pageToken = page[counter]
			previous(counter)
			fmt.Printf("page %d", counter)
		default:
			println("Please check your input or type \"h\" get help")
		}

	}
}

//------------
func startSrv(scope string) *drive.Service {

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	// config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
	// config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/drive")
	config, err := google.ConfigFromJSON(b, scope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)
	// client.Get(url string)
	// srv, err := drive.New(client)
	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}
	return srv
}

// list files of current directory
func list(cmds []string) {

	// parameter setting
	// -dir list files of folder
	// -a show all type of items
	// -d show all folder
	// -l show linked folder
	// -s show started folder
	// -t use file type to filter result
	// -n show by name
	if len(cmds) >= 2 {
		switch cmds[1] {
		case "-dir", "--dir":
			if len(cmds) == 3 {
				qString = "'" + cmds[2] + "' in parents"
			}
			counter = 0
			clearMap()
			userQuery()
		case "-d", "--d":
			qString = "mimeType = 'application/vnd.google-apps.folder'"
			counter = 0
			clearMap()
			userQuery()
		case "-l", "--l":
			qString = "mimeType = 'application/vnd.google-apps.shortcut'"
			counter = 0
			clearMap()
			userQuery()
		case "-s", "--s":
			qString = "starred"
			counter = 0
			clearMap()
			userQuery()
		case "-t", "--t":
			if len(cmds) == 3 {
				qString = "mimeType = '" + cmds[2] + "'"
			}
			counter = 0
			clearMap()
			userQuery()
		case "-n", "--n":
			if len(cmds) == 3 {
				qString = "name contains '" + cmds[2] + "'"
			}
			counter = 0
			clearMap()
			userQuery()
		default:
			qString = ""
			println("this is all ", qString)
			counter = 0
			clearMap()
			userQuery()
		}
	} else {
		counter = 0
		clearMap()
		userQuery()
	}
}

// clear page map for new query
// clearMap ...
func clearMap() {
	for k := range page {
		delete(page, k)
	}
}

// generate prompt suggest for floder
func getSugInfo() func(folder prompt.Suggest) *[]prompt.Suggest {
	a := make([]prompt.Suggest, 0)
	return func(folder prompt.Suggest) *[]prompt.Suggest {
		a = append(a, folder)
		return &a
	}
}


//  generate folder path...
func pathGenerate()  {
	pathInfo := getSugInfo()
	cmd := exec.Command("tree", "-f", "-L", "3", "-i", "-d", os.Getenv("HOME"))
	output, _ := cmd.Output()
	for _, m := range strings.Split(string(output), "\n") {
		// fmt.Printf("metric: %s\n", m)
		s := prompt.Suggest{Text: m, Description: ""}
		pathSug = pathInfo(s)
	}
}


// print the request result
func showResult(counter int, scope string) *drive.FileList {
	// This should testing by change the authorize token
	// r, err := startSrv("https://www.googleapis.com/auth/drive.photos.readonly").Files.List().
	// Spaces("drive").
	// Q("mimeType = 'application/vnd.google-apps.shortcut' or starred").
	// Q("starred").Q("name='IMG_0004.JPG'").
	// Q("starred or name='IMG_0004.JPG'").
	// OrderBy(condition).
	// Corpora("default").
	// fmt.Println("Result start: ", page[counter], qString, counter, scope)
	colorGreen := "\033[32m%26s  %s\t%s\t%s\t%s\n"
	colorCyan := "\033[36m%26s  %s\t%s\t%s\t%s\n"
	dirInfo := getSugInfo()
	fileInfo := getSugInfo()

	r, err := startSrv(scope).Files.List().
		Q(qString).
		PageSize(40).
		Fields("nextPageToken, files(id, name, mimeType, owners, createdTime)").
		PageToken(page[counter]).
		OrderBy("modifiedTime").
		Do()

	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Println("Files:")
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {
			if i.MimeType == "application/vnd.google-apps.folder" {
				// fmt.Println(string(colorGreen), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.CreatedTime)
				fmt.Printf(string(colorGreen), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.CreatedTime)
				s := prompt.Suggest{Text: i.Id, Description: i.Name}
				dirSug = dirInfo(s)
			} else {
				fmt.Printf(string(colorCyan), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.CreatedTime)
				s := prompt.Suggest{Text: i.Id, Description: i.Name}
				fileSug = fileInfo(s)
			}
		}
	}
	return r
}

// rm ... delete file
func rm() {
	//TODO: delete file
	println("this is rm")
}

// WriteCounter counts the number of bytes written to it. It implements to the io.Writer interface
// and we can pass this into io.TeeReader() which will report progress on each write cycle.
type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

// download file
func download(cmds []string) error {
	//TODO: download file
	if len(cmds) >= 2 {
		// println("this is download xx", cmds[1], cmds[2])
		// drive.DriveReadonlyScope
		fgc := startSrv(drive.DriveScope).Files.Get(cmds[1])
		fgc.Header().Add("alt", "media")
		resp, err := fgc.Download()
		// resp, err := startSrv(drive.DriveReadonlyScope).Files.Get("").Header().Add("", "")
		// .Get(cmds[1]).Download(opts ...googleapi.CallOption)
		// resp, err := startSrv(drive.DriveReadonlyScope).Files. 
		// 	Get(cmds[1]).
		// 	Download()

		println("this is download x0")
		if err != nil {
			println("this is download x0" , err.Error())
			return err
			// log.Fatalf("Unable to retrieve files: %v", err)
		}
		println("this is download x1")
		defer resp.Body.Close()
		// Create the file, but give it a tmp file extension, this means we won't overwrite a
		// file until it's downloaded, but we'll remove the tmp extension once downloaded.
		out, err := os.Create(cmds[2]+ ".tmp")
		if err != nil {
			return err
		}
		println("this is download x2")
		// Create our progress reporter and pass it to be used alongside our writer
		counter := &WriteCounter{}
		if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
			out.Close()
			return err
		}
		println("this is download x3")
		// The progress use the same line so print a new line once it's finished downloading
		fmt.Print("\n")

		// Close the file without defer so it can happen before Rename()
		out.Close()

		if err = os.Rename(cmds[2]+".tmp", cmds[2]); err != nil {
			return err
			// log.Fatalf("Unable to save files: %v", err)
		}
		println("this is download x4")
	}
	 return nil
}

// move file
func move() {
	//TODO: move file
	println("this is .. move")
}


// base query
// name ...
func userQuery() {
	r := showResult(counter, drive.DriveScope)
	pageToken = r.NextPageToken
}

// show next page
func next(counter int) {
	r := showResult(counter, drive.DriveScope)
	pageToken = r.NextPageToken
}

// show previous page
func previous(counter int) {
	showResult(counter, drive.DriveScope)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		log.Fatalf("Json encode error: %v", err)
	}
}
