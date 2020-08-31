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

func GetLocalPathInfo() func(prompt prompt.Suggest) *[]prompt.Suggest {
	// to prevent duplicate id is file or folder id, tp =0 is for file name. tp =1 is for file id
	a := make([]prompt.Suggest, 0)
	return func(prompt prompt.Suggest) *[]prompt.Suggest {
		a = append(a, prompt)
		return &a
	}
}
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
// generate prompt suggest for floder
func GetSugInfo() func(prompt prompt.Suggest, id string, tp int) *[]prompt.Suggest {
	// to prevent duplicate id is file or folder id, tp =0 is for file name. tp =1 is for file id
	a := make([]prompt.Suggest, 0)
	c := prompt.Suggest{Text: "", Description: ""}
	return func(prompt prompt.Suggest, id string, tp int) *[]prompt.Suggest {
		a = append(a, prompt)
		// fmt.Printf("metric: %d\n", len(a))
		//-------delete by id
		for i, v := range a {
			if tp == 0 {
				if v.Description == id {
					a[i] = a[len(a)-1]
					a[len(a)-1] = c
					a = a[:len(a)-1]
				}
			}
			if tp == 1 {
				if v.Text == id {
					a[i] = a[len(a)-1]
					a[len(a)-1] = c
					a = a[:len(a)-1]
				}
			}
		}

		//-------clean duplicate
		a = UniquePrompt(a, tp)
		// fmt.Printf("metric: %d\n", len(a))
		return &a
	}
}

func UniquePrompt(pSlice []prompt.Suggest, tp int) []prompt.Suggest {
	// to prevent duplicate id is file or folder id, tp =0 is for file name. tp =1 is for file id
	keys := make(map[string]bool)
	list := []prompt.Suggest{}
	if tp == 0 {
		for _, entry := range pSlice {
			if _, value := keys[entry.Description]; !value {
				keys[entry.Description] = true
				list = append(list, entry)
			}
		}
	}
	if tp == 1 {
		for _, entry := range pSlice {
			if _, value := keys[entry.Text]; !value {
				keys[entry.Text] = true
				list = append(list, entry)
			}
		}

	}
	return list
}

// Unique int slice
func Unique(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// for comment function to generate gmail & domain prompt
func LoadproSugg(fileName string) *[]prompt.Suggest {
	sug := GetSugInfo()
	var ssug *[]prompt.Suggest
	mail, err := ioutil.ReadFile(GetAppHome() + string(os.PathSeparator) + fileName)
	if err != nil {
		glog.V(5).Info(err.Error())
		p := prompt.Suggest{Text: "", Description: ""}
		ssug = sug(p, "", 0)
		return ssug
	}
	mdata := []prompt.Suggest{}
	err = json.Unmarshal([]byte(mail), &mdata)
	if err != nil {
		// glog.Warning(err.Error())
		glog.V(5).Info(err.Error())
		p := prompt.Suggest{Text: "", Description: ""}
		ssug = sug(p, "", 0)
		return ssug
	}
	for _, value := range mdata {
		p := prompt.Suggest{Text: value.Text, Description: value.Description}
		ssug = sug(p, "", 0)
	}

	return ssug
}

func SaveProperty(name string, v interface{}) {
	file, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		glog.Error("JSON Marshal error: ", err.Error())
	}
	err = ioutil.WriteFile(GetAppHome()+string(os.PathSeparator)+name+".json", file, 0644)
	if err != nil {
		glog.Error("Write File error: ", err.Error())
	}
}

func GetAppHome() string {
	home, _ := os.UserHomeDir()
	path := fmt.Sprint(home, string(os.PathSeparator), ".local", string(os.PathSeparator), "goclue")
	// glog.V(8).Info(path)
	if !Exists(path) {
		glog.V(8).Info(path, " Not exist")
		if err := os.MkdirAll(path, 0755); err != nil {
			glog.Error("Create app folder failed: ", err.Error())
			return ""
		}
	}
	return path
}

func CheckCredentials(fail Callback, success Callback) {
	home, _ := os.UserHomeDir()
	path := fmt.Sprint(home, string(os.PathSeparator),
		".local", string(os.PathSeparator),
		"goclue", string(os.PathSeparator),
		"credentials.json")
	// glog.V(8).Info(path)
	if Exists(path) {
		success()
	} else {
		fail()
	}
}

type Callback func()

func Check(path string, fail Callback, success Callback) bool {
	if Exists(path) {
		success()
		return true
	} else {
		fail()
		return false
	}
}

func Movefile(from, to string) bool {
	err := os.Rename(from, to)
	return err == nil
	// success()
}

// determine array contain string
func IsContain(items []prompt.Suggest, item string) bool {
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
	return func(path, id, file string) map[string]string {
		// files = append(files, path+string(os.PathSeparator)+file)
		files[id] = path + string(os.PathSeparator) + file
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
	glog.V(8).Info("GetFilesAndFolders end: ", "fis size: ", len(files), " fods size: ", len(folders))

	return files, folders, nil
}

func StartSrv(scope string) *drive.Service {

	b, err := ioutil.ReadFile(GetAppHome() + string(os.PathSeparator) + "credentials.json")
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
	tokFile := GetAppHome() + string(os.PathSeparator) + "token.json"
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
