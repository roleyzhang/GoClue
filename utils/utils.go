package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/golang/glog"

	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// var (
// // counter int
// )

// func init() {
// }

// var patf string

// func incrFiles() func(path, file string) []string {
// 	var files []string
// 	return func(path, file string) []string {
// 		files = append(files, path+string(os.PathSeparator)+file)
// 		return files
// 	}
// }

// determine array contain string
func IsContain(items []prompt.Suggest , item string) bool {
	for _, eachItem := range items {
		if strings.Contains(eachItem.Text, item) {
			return true
		}
	}
	return false
}

// Checking linux system commands 
func IsCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
    if _, err := os.Stat(name); err != nil {
        if os.IsNotExist(err) {
            return false
        }
    }
    return true
}

// checking whether is folder
func IsDir(path string) bool {
	// glog.V(8).Info(path)
    s, err := os.Stat(path)
    if err != nil {
		// glog.V(8).Info(err)
        return false
    }
    return s.IsDir()
}

// checking whether is file
func IsFile(path string) bool {
    return !IsDir(path)
}

// clearMap ...
func ClearDownloadMap(m map[string]string) {
	for k := range m {
		delete(m, k)
	}
}

func IncrFiles() func(path, id, file string) map[string]string {
	var files = make(map[string]string) 
	return func(path, id, file string) map[string]string  {
		// files = append(files, path+string(os.PathSeparator)+file)
		files[id] = path+string(os.PathSeparator)+file
		return files
	}
}
var filesFromSrv = IncrFiles()

func GetFilesAndFolders(id, path string) (files map[string]string, folders []string, err error) {
	pthSep := string(os.PathSeparator)
	qString := "'" + id + "' in parents"
	// glog.V(8).Info("GetFilesAndFolders qString: ", pthSep, qString)
	item, err := StartSrv(drive.DriveScope).Files.List().
		Q(qString).PageSize(40). //"nextPageToken, files(id, name, mimeType, parents)")
		Fields("nextPageToken, files(id, name, mimeType)").
		Do()
	if err != nil {
		glog.Errorln("file or dir not exist: ", err.Error())
		return nil, nil, err
	}
	glog.V(8).Info("GetFilesAndFolders item len: ", len(item.Files))
	for _, file := range item.Files {
		if file.MimeType == "application/vnd.google-apps.folder" {
			pat := path + pthSep + file.Name
			glog.V(8).Info("D: ", pat)
			folders = append(folders, pat)
			_, _, err := GetFilesAndFolders(file.Id, pat)
			if err != nil {
				glog.Errorln("file or dir not exist: ", err.Error())
				return nil, nil, err
			}
		} else {
			files = filesFromSrv(path, file.Id, file.Name)
			glog.V(8).Info("F: ", path+pthSep+file.Name)
			// files = append(files, path+pthSep+file.Name)
		}
	}
	glog.V(8).Info("GetFilesAndFolders end: " ,"fis size: ",len(files)," fods size: ", len(folders))

	return files, folders, nil
}

func StartSrv(scope string) *drive.Service {

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		glog.Errorf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	// config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
	// config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/drive")
	config, err := google.ConfigFromJSON(b, scope)
	if err != nil {
		glog.Errorf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)
	// client.Get(url string)
	// srv, err := drive.New(client)
	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		glog.Errorf("Unable to retrieve Drive client: %v", err)
	}
	return srv
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
		glog.Errorf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		glog.Errorf("Unable to retrieve token from web %v", err)
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
		glog.Errorf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		glog.Errorf("Json encode error: %v", err)
	}
}
