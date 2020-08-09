package cmd


import (
	"fmt"
	"net/http"
	"io/ioutil"
	"log"
	"errors"
	"strings"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	// "os/exec"
	"encoding/json"
	"golang.org/x/oauth2"
	"github.com/c-bata/go-prompt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)


var qString string
var dirSug *[]prompt.Suggest
var page map[int]string
var fileSug *[]prompt.Suggest
var pathSug *[]prompt.Suggest
var allSug *[]prompt.Suggest
var idfileSug *[]prompt.Suggest
var iddirSug *[]prompt.Suggest
var colorGreen string
var colorCyan string
var colorYellow string
var colorRed string

var LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}


type ItemInfo struct {
	// item       *drive.File
	Path   map[string]string
	RootId string
	ItemId string
}

func init() {
	colorGreen = "\033[32m%26s  %s\t%s\t%s\t%s\n"
	colorCyan = "\033[36m%26s  %s\t%s\t%s\t%s\n"
	colorYellow = "\033[33m%s %s %s\n"
	colorRed = "\033[31m%s\n"
	page = make(map[int]string)
}

// getSugId ...
func getSugId(sug *[]prompt.Suggest, text string) (string, error) {
	// fmt.Println(text)
	if sug != nil {
		for _, v := range *sug {
			if v.Text == text {
				return v.Description, nil
			}
		}
	}
	qString := "name='" + text + "'"+ " and trashed=false"

	file, err := startSrv(drive.DriveScope).Files.List().
		Q(qString).
		PageSize(2).
		Fields("nextPageToken, files(id, name, mimeType, owners, createdTime)").
		// Fields("id, name, mimeType, parents, createdTime").
		Do()

	if err != nil {
		fmt.Printf(string(colorRed), err.Error())
		return "", err
	}
	// fmt.Printf(string(colorRed), len(file.Files))
	if len(file.Files) > 1 {
		fmt.Printf(string(colorRed), "The file name is not unique, please use file id do the operation")
		return "", nil
	}

	if len(file.Files) == 0 {
		return "", errors.New("No Item has been found")
	}
	return file.Files[0].Id, nil
}

// generate prompt suggest for floder
func getSugInfo() func(folder prompt.Suggest) *[]prompt.Suggest {
	a := make([]prompt.Suggest, 0)
	return func(folder prompt.Suggest) *[]prompt.Suggest {
		a = append(a, folder)
		return &a
	}
}

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

// msg ...
func msg(message string) {
	LivePrefixState.LivePrefix = message + ">>> "
	LivePrefixState.IsEnable = true
}

// getSugDec ...
func getSugDec(sug *[]prompt.Suggest, text string) string {
	if sug != nil {
		for _, v := range *sug {
			if v.Description == text {
				// fmt.Println(v.Description)
				// return v.Description
				return v.Text
			}
		}
	} else {
		return text
	}
	return ""
}

// breakDown ...
func breakDown(path string) []string {
	return strings.Split(path, "/")
}

// print the request result
func (ii *ItemInfo)ShowResult(counter int, scope string) *drive.FileList {
	// This should testing by change the authorize token
	// r, err := startSrv("https://www.googleapis.com/auth/drive.photos.readonly").Files.List().
	// Spaces("drive").
	// Q("mimeType = 'application/vnd.google-apps.shortcut' or starred").
	// Q("starred").Q("name='IMG_0004.JPG'").
	// Q("starred or name='IMG_0004.JPG'").
	// OrderBy(condition).
	// Corpora("default").
	// fmt.Println("Result start: ", page[counter], qString, counter, scope)
	// colorGreen := "\033[32m%26s  %s\t%s\t%s\t%s\n"
	// colorCyan := "\033[36m%26s  %s\t%s\t%s\t%s\n"
	dirInfo := getSugInfo()
	fileInfo := getSugInfo()
	allInfo := getSugInfo()
	idfileInfo := getSugInfo()
	iddirInfo := getSugInfo()

	//--------every time runCommand add folder history to dirSug
	for key, value := range ii.Path {
		s := prompt.Suggest{Text: value, Description: key}
		dirSug = dirInfo(s)

	}

	fmt.Println("qString:", qString)
	r, err := startSrv(scope).Files.List().
		Q(qString).
		PageSize(40).
		Fields("nextPageToken, files(id, name, mimeType, owners, parents, createdTime)").
		PageToken(page[counter]).
		// OrderBy("modifiedTime").
		Do()

	if err != nil {
		fmt.Printf(string(colorRed), "Unable to retrieve files: %v", err.Error())
		// log.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Println("Files:")
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {
			if i.MimeType == "application/vnd.google-apps.folder" {
				// fmt.Println(string(colorGreen), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.CreatedTime)
				// fmt.Printf(string(colorGreen), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.Parents, i.CreatedTime)
				fmt.Printf(string(colorGreen), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.CreatedTime)
				s := prompt.Suggest{Text: i.Id, Description: i.Name}
				s2 := prompt.Suggest{Text: i.Name, Description: i.Id}
				dirSug = dirInfo(s2)
				allSug = allInfo(s2)
				iddirSug = iddirInfo(s)
				// 	s := prompt.Suggest{Text: i.Id, Description: i.Name}
				// 	dirSug = dirInfo(s)
			} else {
				// fmt.Printf(string(colorCyan), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.Parents, i.CreatedTime)
				fmt.Printf(string(colorCyan), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.CreatedTime)
				s := prompt.Suggest{Text: i.Id, Description: i.Name}
				s2 := prompt.Suggest{Text: i.Name, Description: i.Id}
				fileSug = fileInfo(s2)
				allSug = allInfo(s2)
				idfileSug = idfileInfo(s)
			}
		}
	}
	return r
}
//  generate folder path...
func PathGenerate(path string) {
	pathInfo := getSugInfo()
	if path == "HOME" {
		cmd := exec.Command("tree", "-f", "-L", "3", "-i", "-d", os.Getenv(path))
		output, _ := cmd.Output()
		for _, m := range strings.Split(string(output), "\n") {
			// fmt.Printf("metric: %s\n", m)
			s := prompt.Suggest{Text: m, Description: ""}
			pathSug = pathInfo(s)
		}
	} else {
		cmd := exec.Command("tree", "-f", "-L", "3", "-i", path)
		output, _ := cmd.Output()
		for _, m := range strings.Split(string(output), "\n") {
			// fmt.Printf("metric: %s\n", m)
			s := prompt.Suggest{Text: m, Description: ""}
			pathSug = pathInfo(s)
		}

	}
}
// SetPrefix ...
func (ii *ItemInfo) SetPrefix(msgs string) {
	// folderId := ii.path[len(ii.path)-1]
	// fmt.Println(ii.itemId)
	folderId := ii.ItemId
	if dirSug != nil {
		folderName := getSugDec(dirSug, folderId)
		msg(folderName + msgs)
	}
}

// rmd ... delete file by id
func (ii *ItemInfo) Rmd(id, types string) error {
	//TODO: delete file
		file, err := startSrv(drive.DriveScope).Files.Get(id).Do()
		if err != nil {
			log.Println("file or dir not exist: " + err.Error())
			return err
		}

		if id == ii.RootId {
			return errors.New("The root folder should not be deleted")
		}

		if types == "-r" && file.MimeType != "application/vnd.google-apps.folder" {
			return errors.New("The delete item: item is not folder")
		}

		err = startSrv(drive.DriveScope).Files.Delete(id).Do()

		if err != nil {
			log.Println("file or dir delete failed: " + err.Error())
			return err
		}
		return nil
}

// rm ... delete file
func (ii *ItemInfo) Rm(name, types string) error {
	//TODO: delete file
	var id string
	iD, err := getSugId(dirSug, strings.TrimSuffix(name, " "))
	if err != nil {
		fmt.Printf(string(colorRed), "file or dir not exist: "+err.Error())
		log.Println("file or dir not exist: " + err.Error())
		return err
	}
	id = iD

	file, err := startSrv(drive.DriveScope).Files.Get(id).Do()
	if err != nil {
		log.Println("file or dir not exist: " + err.Error())
		return err
	}

	if id == ii.RootId {
		return errors.New("The root folder should not be deleted")
	}

	if types == "-r" && file.MimeType != "application/vnd.google-apps.folder" {
		return errors.New("The delete item: item is not folder")
	}

	err = startSrv(drive.DriveScope).Files.Delete(id).Do()

	if err != nil {
		log.Println("file or dir delete failed: " + err.Error())
		return err
	}
	return nil
}

// trash ...
func (ii *ItemInfo) Trash(name, types string) error {
	//TODO: trash file
	var id string
	iD, err := getSugId(dirSug, strings.TrimSuffix(name, " "))
	if err != nil {
		fmt.Printf(string(colorRed), "file or dir not exist: "+err.Error())
		log.Println("file or dir not exist: " + err.Error())
		return err
	}
	id = iD

	file, err := startSrv(drive.DriveScope).Files.Get(id).Do()
	if err != nil {
		log.Println("file or dir not exist: " + err.Error())
		return err
	}

	if id == ii.RootId {
		return errors.New("The root folder should not be trashed")
	}

	if types == "-r" && file.MimeType != "application/vnd.google-apps.folder" {
		return errors.New("The trashed item: item is not folder")
	}

	_, err = startSrv(drive.DriveScope).Files.Update(file.Id, &drive.File{Trashed: true}).Do()

	if err != nil {
		log.Println("file or dir trashed failed: " + err.Error())
		return err
	}
	fmt.Printf(string(colorYellow), file.Name, "", "Be Trashed")
	return nil
}

// trash by id...
func (ii *ItemInfo) Trashd(id, types string) error {
	file, err := startSrv(drive.DriveScope).Files.Get(id).Do()
	if err != nil {
		log.Println("file or dir not exist: " + err.Error())
		return err
	}

	if id == ii.RootId {
		return errors.New("The root folder should not be trashed")
	}

	if types == "-r" && file.MimeType != "application/vnd.google-apps.folder" {
		return errors.New("The trashed item: item is not folder")
	}

	file, err = startSrv(drive.DriveScope).Files.Update(file.Id, &drive.File{Trashed: true}).Do()

	if err != nil {
		log.Println("file or dir trashed failed: " + err.Error())
		return err
	}
	fmt.Printf(string(colorYellow), file.Name, "", "Be Trashed")
	return nil
}

func (ii *ItemInfo) Upload(file string) (*drive.File, error) {
	fil := strings.Split(file, "u ")
	// fmt.Println(fil)
	fi, err := os.Open(fil[1])
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	fileInfo, err := fi.Stat()
	if err != nil {
		return nil, err
	}
	u := &drive.File{
		Name:    filepath.Base(fileInfo.Name()),
		Parents: []string{ii.ItemId},
	}
	// ufile, err := startSrv(drive.DriveScope).Files.Create(u).Media(fi).Do()
	ufile, err := startSrv(drive.DriveScope).Files.
		Create(u).
		ResumableMedia(context.Background(), fi, fileInfo.Size(), "").
		ProgressUpdater(func(now, size int64) { fmt.Printf("%d, %d\r", now, size) }).
		Do()
	if err != nil {
		return nil, err
	}
	return ufile, nil
}

// createDir...
func (ii *ItemInfo) CreateDir(name string) (*drive.File, error) {
	d := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{ii.ItemId},
	}
	dir, err := startSrv(drive.DriveScope).Files.Create(d).Do()

	if err != nil {
		log.Println("Could not create dir: " + err.Error())
		return nil, err
	}

	fmt.Printf(string(colorYellow), dir.Name, "", " has been created")
	return dir, nil
}

// getRoot ...
func (ii *ItemInfo) GetRoot() {
	dirInfo := getSugInfo()
	item, err := startSrv(drive.DriveScope).
		// Files.Get(id).
		Files.Get("root").
		Fields("id, name, mimeType, parents, owners, createdTime").
		Do()
	if err != nil {
		println("shit happened: ", err.Error())
		log.Fatalf("Unable to retrieve root: %v", err)
		// return nil
	}
	if item.MimeType == "application/vnd.google-apps.folder" {
		ii.Path[item.Id] = item.Name
		ii.ItemId = item.Id
		ii.RootId = item.Id
		// setting the prompt.Suggest
		s2 := prompt.Suggest{Text: item.Name, Description: item.Id}
		dirSug = dirInfo(s2)
	}
}

// getNode by id ...
func (ii *ItemInfo) GetNoded(id string) {
	// println(id)
	item, err := startSrv(drive.DriveScope).
		Files.Get(id).
		// Files.Get("root").
		Fields("id, name, mimeType, parents, owners, createdTime").
		Do()
	if err != nil {
		println("shit happened: ", err.Error())
		log.Fatalf("Unable to retrieve root: %v", err)
		// return nil
	}
	if item.MimeType == "application/vnd.google-apps.folder" || item.MimeType == "application/vnd.google-apps.shortcut" {
		ii.Path[item.Id] = item.Name
		ii.ItemId = item.Id
		if id == "root" {
			ii.RootId = item.Id
		}
	}
}

// getNode ...
func (ii *ItemInfo) GetNode(cmd string) {
	// println(id)
	var id string

	name := strings.Trim(strings.Split(cmd, "cd ")[1], " ")
	dirInfo := getSugInfo()

	if name == "root" || name == "My Drive" {
		id = "root"
	} else {
		iD, err := getSugId(dirSug, strings.TrimSuffix(name, " "))
		if err != nil {
			fmt.Printf(string(colorRed), "file or dir not exist: "+err.Error())
			log.Println("file or dir not exist: " + err.Error())
			// return nil, err
		}
		id = iD
	}
	item, err := startSrv(drive.DriveScope).
		Files.Get(id).
		// Files.Get("root").
		Fields("id, name, mimeType, parents, owners, createdTime").
		Do()
	if err != nil {
		println("shit happened: ", err.Error())
		log.Fatalf("Unable to retrieve root: %v", err)
		// return nil
	}
	if item.MimeType == "application/vnd.google-apps.folder" || item.MimeType == "application/vnd.google-apps.shortcut" {
		ii.Path[item.Id] = item.Name
		ii.ItemId = item.Id
		if id == "root" {
			ii.RootId = item.Id
		}

		// setting the prompt.Suggest
		s2 := prompt.Suggest{Text: item.Name, Description: item.Id}
		dirSug = dirInfo(s2)
	}
}

// move file
func (ii *ItemInfo) Move(cmd string) error {
	//TODO: move file
	// println("this is .. move", cmd)
	if !strings.Contains(cmd, ">") {
		fmt.Printf(string(colorRed), "Wrong command format, please use \"h\" get help")
		return errors.New("Wrong command format, please use \"h\" get help")
	}
	fil := strings.Split(strings.Split(cmd, "mv ")[1], ">")
	iD, err := getSugId(allSug, strings.TrimSuffix(fil[0], " "))
	if err != nil {
		log.Println("file or dir not exist: " + err.Error())
		return err
	}
	file, err := startSrv(drive.DriveScope).Files.Get(iD).Fields("id, name, mimeType, parents, createdTime").Do()
	if err != nil {
		log.Println("file or dir not exist: " + err.Error())
		return err
	}

	if file.Id == ii.RootId {
		return errors.New("The root folder should not be moved")
	}

	if len(breakDown(fil[1])) > 1 { // move to another folder
		newParentName := strings.Trim(breakDown(fil[1])[len(breakDown(fil[1]))-2], " ") // move to another folder
		newName := strings.Trim(breakDown(fil[1])[len(breakDown(fil[1]))-1], " ")       // change item name
		iD, err := getSugId(allSug, newParentName)
		if err != nil {
			log.Println("file or dir not exist: " + err.Error())
			return err
		}

		var parents string
		if len(file.Parents) > 0 {
			parents = file.Parents[0]
		}
		newFile, err := startSrv(drive.DriveScope).Files.Update(file.Id, &drive.File{
			Name: newName,
		}).AddParents(iD).
			RemoveParents(parents).Do()
		if err != nil {
			fmt.Printf(string(colorRed), err.Error())
			return err
		}
		fmt.Printf(string(colorYellow), file.Name, "->", path.Join(newParentName, newFile.Name))
	} else { // change file name
		newName := strings.Trim(breakDown(fil[1])[len(breakDown(fil[1]))-1], " ") // change item name
		newFile, err := startSrv(drive.DriveScope).Files.Update(file.Id, &drive.File{
			Name: newName,
		}).Do()
		// .AddParents(file.Id)
		// RemoveParents(path.Join(file.item.Parents...)).
		// Fields(fileInfoFields...).Do()
		if err != nil {
			fmt.Printf(string(colorRed), err.Error())
			return err
		}
		fmt.Printf(string(colorYellow), file.Name, "->", newFile.Name)
	}
	// }
	return nil
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
