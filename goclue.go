package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

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

type command struct {
	name  string
	param string
	tip   string
}

var allCommands []command
var pageToken string
var counter int
var page map[int]string
// var service *drive.Service

func init() {
	fmt.Println("This will get called on main initialization")
	// allCommands = make([]command, 0)
	allCommands = []command{
		{"q", "", "Quit"},
		{"login", "", "Login to your account of net drive"},
		{"mkdir", "", "Create directory"},
		{"rm", "", "Delete directory or file, use \"-r\" for delete directory"},
		{"cd", "", "change directory"},
		{"..", "", "Exit current directory"},
		{"d", "", "Download files use \"-r\" for download directory"},
		{"ls", "", "list contents of current directory"},
		{"u", "", "Upload directory or file, use \"-r\" for upload directory"},
		{"h", "", "Print help"},
		{"n", "", "Next page"},
		{"p", "", "Previous page"},
	}

	page = make(map[int]string)
}

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
		case "..":
			println("this is ..")
		case "rm":
			println("this is rm")
		case "d":
			println("this is download")
		case "ls":
			list()
			println("this is ls")
		case "u":
			println("this is upload")
		case "h":
			for _, cmd := range allCommands {
				fmt.Printf("%6s: %s %s \n", cmd.name, cmd.param, cmd.tip)
			}
		case "n":
			counter++
			fmt.Printf("this is next page %d", counter)
			if page[counter] == "" {
				page[counter] = pageToken
			}
			next(counter)
		case "p":
			if counter > 0 {
				counter--
			}
			pageToken = page[counter]
			previous(counter)
			fmt.Printf("this is previous page %d", counter)
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

	// srv, err := drive.New(client)
	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}
	return srv
}

// list files of current directory
func list() {

	// parameter setting
	// -a show all type of items
	// -d show all folder
	// -l show linked folder
	// -s show started folder
	// r, err := srv.Files.List().
	colorGreen := "\033[32m"
	// colorCyan := "\033[36m"
	r, err := startSrv(drive.DriveMetadataReadonlyScope).Files.List().
		PageSize(200).
		Fields("nextPageToken, files(id, name, mimeType)").
		PageToken(pageToken).
		OrderBy("starred, createdTime").
		// Corpora("default").
		Do()

	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Println("Files:")
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {
			if  i.MimeType == "application/vnd.google-apps.folder" {
				// fmt.Printf("%s (%s)\n", i.Name, i.Id)
				fmt.Println(string(colorGreen), i.Name, i.Id, i.MimeType )
			}

			// else{
			// 	fmt.Println(string(colorCyan), i.Name, i.Id, i.MimeType)

			// }
			// fmt.Printf("%s (%s) %s \n", i.Name, i.Id, i.MimeType)
		}
	}
	pageToken = r.NextPageToken

	// r, err := startSrv(drive.DriveMetadataReadonlyScope).Files.Get("labels/starred").Do()
}

// show next page
func next(counter int) {
	// b, err := ioutil.ReadFile("credentials.json")
	// if err != nil {
	// 	log.Fatalf("Unable to read client secret file: %v", err)
	// }

	// // If modifying these scopes, delete your previously saved token.json.
	// // config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
	// config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/drive")
	// if err != nil {
	// 	log.Fatalf("Unable to parse client secret file to config: %v", err)
	// }
	// client := getClient(config)

	// // srv, err := drive.New(client)
	// ctx := context.Background()
	// srv, err := drive.NewService(ctx, option.WithHTTPClient(client))

	// if err != nil {
	// 	log.Fatalf("Unable to retrieve Drive client: %v", err)
	// }

	colorGreen := "\033[32m"
	colorCyan := "\033[36m"
	r, err := startSrv(drive.DriveMetadataReadonlyScope).Files.List().PageSize(20).
		// r, err := srv.Files.List().
		Fields("nextPageToken, files(id, name, mimeType)").PageToken(page[counter]).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Println("Files:")
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {
			// fmt.Printf("%s (%s)\n", i.Name, i.Id)
			if  i.MimeType == "application/vnd.google-apps.folder" {
				// fmt.Printf("%s (%s)\n", i.Name, i.Id)
				fmt.Println(string(colorGreen), i.Name, i.Id)
			}else{
				fmt.Println(string(colorCyan), i.Name, i.Id)

			}
		}
	}
	pageToken = r.NextPageToken
}

// show previous page
func previous(counter int) {
	// b, err := ioutil.ReadFile("credentials.json")
	// if err != nil {
	// 	log.Fatalf("Unable to read client secret file: %v", err)
	// }

	// // If modifying these scopes, delete your previously saved token.json.
	// // config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
	// config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/drive")
	// if err != nil {
	// 	log.Fatalf("Unable to parse client secret file to config: %v", err)
	// }
	// client := getClient(config)

	// // srv, err := drive.New(client)
	// ctx := context.Background()
	// srv, err := drive.NewService(ctx, option.WithHTTPClient(client))

	// if err != nil {
	// 	log.Fatalf("Unable to retrieve Drive client: %v", err)
	// }

	colorGreen := "\033[32m"
	colorCyan := "\033[36m"
	r, err := startSrv(drive.DriveMetadataReadonlyScope).Files.List().PageSize(20).
		// r, err := srv.Files.List().
		Fields("nextPageToken, files(id, name, mimeType)").PageToken(page[counter]).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Println("Files:")
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {
			// fmt.Printf("%s (%s)\n", i.Name, i.Id)
			if  i.MimeType == "application/vnd.google-apps.folder" {
				// fmt.Printf("%s (%s)\n", i.Name, i.Id)
				fmt.Println(string(colorGreen), i.Name, i.Id)
			}else{
				fmt.Println(string(colorCyan), i.Name, i.Id)

			}
		}
	}
	// pageToken = r.NextPageToken
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
