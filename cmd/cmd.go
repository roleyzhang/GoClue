package cmd

import (
	"errors"
	"github.com/golang/glog"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"fmt"
	// "os/exec"
	"golang.org/x/net/context"
	"github.com/c-bata/go-prompt"
	"github.com/dustin/go-humanize"
	"github.com/roleyzhang/GoClue/utils"
	"google.golang.org/api/drive/v3"
)

var qString string

var DirSug *[]prompt.Suggest
var Page map[int]string
var FileSug *[]prompt.Suggest
var PathSug *[]prompt.Suggest
var AllSug *[]prompt.Suggest
var IdfileSug *[]prompt.Suggest
var IddirSug *[]prompt.Suggest

var colorGreen string
var colorCyan string
var colorYellow string
var colorRed string

var commands map[string]string

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

	// for list function
	commands = make(map[string]string)
	commands["default"] = "trashed=false"
	commands["dls"] = "'$' in parents and trashed=false"
	commands["dir"] = "'$' in parents"
	commands["d"] = "mimeType = 'application/vnd.google-apps.folder' and trashed=false"
	commands["l"] = "mimeType = 'application/vnd.google-apps.shortcut'"
	commands["s"] = "starred"
	commands["t"] = "mimeType = '$' and trashed=false"
	commands["n"] = "name contains '$' and trashed=false"
	commands["tr"] = "trashed=true"
	commands["c"] = "fullText contains '$' and trashed=false"
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
	qString := "name='" + text + "'" + " and trashed=false"

	file, err := utils.StartSrv(drive.DriveScope).Files.List().
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


// breakDown ...
func breakDown(path string) []string {
	return strings.Split(path, "/")
}

// print the request result
func (ii *ItemInfo) ShowResult(
	page map[int]string,
	counter int,
	param, cmd, scope string) *drive.FileList {
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
		DirSug = dirInfo(s)

	}
	if param != "next" && param != "previous" {
		qString = commands[param]
		if strings.Contains(qString, "$") {
			qString = strings.ReplaceAll(qString, "$", cmd)
		}
	}
	if param == "dir" {
		iD, err := getSugId(AllSug, cmd)
		if err != nil {
			fmt.Printf(string(colorRed), "file or dir not exist: "+err.Error())
			glog.Errorln("file or dir not exist: " + err.Error())
		}
		qString = commands[param]
		if strings.Contains(qString, "$") {
			qString = strings.ReplaceAll(qString, "$", iD)
		}
	}

	glog.V(5).Infoln("qString: ", qString)
	r, err := utils.StartSrv(scope).Files.List().
		Q(qString).
		PageSize(40).
		Fields("nextPageToken, files(id, name, mimeType, owners, parents, createdTime)").
		PageToken(page[counter]).
		// OrderBy("modifiedTime").
		Do()

	if err != nil {
		fmt.Printf(string(colorRed), "Unable to retrieve files: %v", err.Error())
		// uncomment below will cause 500 error and program exit why?
		// glog.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Println("Files:")
	if len(r.Files) == 0 {
		fmt.Printf(string(colorYellow), "No files found.", "", "")
		// fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {
			if i.MimeType == "application/vnd.google-apps.folder" {
				// fmt.Println(string(colorGreen), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.CreatedTime)
				// fmt.Printf(string(colorGreen), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.Parents, i.CreatedTime)
				fmt.Printf(string(colorGreen), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.CreatedTime)
				s := prompt.Suggest{Text: i.Id, Description: i.Name}
				s2 := prompt.Suggest{Text: i.Name, Description: i.Id}
				DirSug = dirInfo(s2)
				AllSug = allInfo(s2)
				IddirSug = iddirInfo(s)
				// 	s := prompt.Suggest{Text: i.Id, Description: i.Name}
				// 	dirSug = dirInfo(s)
			} else {
				// fmt.Printf(string(colorCyan), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.Parents, i.CreatedTime)
				fmt.Printf(string(colorCyan), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.CreatedTime)
				s := prompt.Suggest{Text: i.Id, Description: i.Name}
				s2 := prompt.Suggest{Text: i.Name, Description: i.Id}
				FileSug = fileInfo(s2)
				AllSug = allInfo(s2)
				IdfileSug = idfileInfo(s)
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
			PathSug = pathInfo(s)
		}
	} else {
		cmd := exec.Command("tree", "-f", "-L", "3", "-i", path)
		output, _ := cmd.Output()
		for _, m := range strings.Split(string(output), "\n") {
			// fmt.Printf("metric: %s\n", m)
			s := prompt.Suggest{Text: m, Description: ""}
			PathSug = pathInfo(s)
		}

	}
}

// SetPrefix ...

// rmd ... delete file by id
func (ii *ItemInfo) Rmd(id, types string) error {
	//TODO: delete file
	file, err := utils.StartSrv(drive.DriveScope).Files.Get(id).Do()
	if err != nil {
		glog.Error("file or dir not exist: ", err.Error())
		return err
	}

	if id == ii.RootId {
		return errors.New("The root folder should not be deleted")
	}

	if types == "-r" && file.MimeType != "application/vnd.google-apps.folder" {
		return errors.New("The delete item: item is not folder")
	}

	err = utils.StartSrv(drive.DriveScope).Files.Delete(id).Do()

	if err != nil {
		glog.Errorln("file or dir delete failed: " + err.Error())
		return err
	}
	return nil
}

// rm ... delete file
func (ii *ItemInfo) Rm(name, types string) error {
	//TODO: delete file
	var id string
	iD, err := getSugId(DirSug, strings.TrimSuffix(name, " "))
	if err != nil {
		fmt.Printf(string(colorRed), "file or dir not exist: "+err.Error())
		glog.Errorln("file or dir not exist: ", err.Error())
		return err
	}
	id = iD

	file, err := utils.StartSrv(drive.DriveScope).Files.Get(id).Do()
	if err != nil {
		glog.Errorln("file or dir not exist: ", err.Error())
		return err
	}

	if id == ii.RootId {
		return errors.New("The root folder should not be deleted")
	}

	if types == "-r" && file.MimeType != "application/vnd.google-apps.folder" {
		return errors.New("The delete item: item is not folder")
	}

	err = utils.StartSrv(drive.DriveScope).Files.Delete(id).Do()

	if err != nil {
		glog.Errorln("file or dir delete failed: ", err.Error())
		return err
	}
	return nil
}

// trash ...
func (ii *ItemInfo) Trash(name, types string) error {
	//TODO: trash file
	var id string
	iD, err := getSugId(DirSug, strings.TrimSuffix(name, " "))
	if err != nil {
		fmt.Printf(string(colorRed), "file or dir not exist: ", err.Error())
		glog.Errorln("file or dir not exist: ", err.Error())
		return err
	}
	id = iD

	file, err := utils.StartSrv(drive.DriveScope).Files.Get(id).Do()
	if err != nil {
		glog.Errorln("file or dir not exist: ", err.Error())
		return err
	}

	if id == ii.RootId {
		return errors.New("The root folder should not be trashed")
	}

	if types == "-r" && file.MimeType != "application/vnd.google-apps.folder" {
		return errors.New("The trashed item: item is not folder")
	}

	_, err = utils.StartSrv(drive.DriveScope).Files.Update(file.Id, &drive.File{Trashed: true}).Do()

	if err != nil {
		glog.Errorln("file or dir trashed failed: ", err.Error())
		return err
	}
	fmt.Printf(string(colorYellow), file.Name, "", "Be Trashed")
	return nil
}

// trash by id...
func (ii *ItemInfo) Trashd(id, types string) error {
	file, err := utils.StartSrv(drive.DriveScope).Files.Get(id).Do()
	if err != nil {
		glog.Errorln("file or dir not exist: ", err.Error())
		return err
	}

	if id == ii.RootId {
		return errors.New("The root folder should not be trashed")
	}

	if types == "-r" && file.MimeType != "application/vnd.google-apps.folder" {
		return errors.New("The trashed item: item is not folder")
	}

	file, err = utils.StartSrv(drive.DriveScope).Files.Update(file.Id, &drive.File{Trashed: true}).Do()

	if err != nil {
		glog.Errorln("file or dir trashed failed: ", err.Error())
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
	ufile, err := utils.StartSrv(drive.DriveScope).Files.
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
	dir, err := utils.StartSrv(drive.DriveScope).Files.Create(d).Do()

	if err != nil {
		glog.Errorln("Could not create dir: ", err.Error())
		return nil, err
	}

	fmt.Printf(string(colorYellow), dir.Name, "", " has been created")
	return dir, nil
}

// getRoot ...
func (ii *ItemInfo) GetRoot() {
	dirInfo := getSugInfo()
	item, err := utils.StartSrv(drive.DriveScope).
		// Files.Get(id).
		Files.Get("root").
		Fields("id, name, mimeType, parents, owners, createdTime").
		Do()
	if err != nil {
		println("shit happened: ", err.Error())
		glog.Fatalf("Unable to retrieve root: %v", err)
		// return nil
	}
	if item.MimeType == "application/vnd.google-apps.folder" {
		ii.Path[item.Id] = item.Name
		ii.ItemId = item.Id
		ii.RootId = item.Id
		// setting the prompt.Suggest
		s2 := prompt.Suggest{Text: item.Name, Description: item.Id}
		DirSug = dirInfo(s2)
	}
}

// getNode by id ...
func (ii *ItemInfo) GetNoded(id string) {
	// println(id)
	item, err := utils.StartSrv(drive.DriveScope).
		Files.Get(id).
		// Files.Get("root").
		Fields("id, name, mimeType, parents, owners, createdTime").
		Do()
	if err != nil {
		println("shit happened: ", err.Error())
		glog.Fatalf("Unable to retrieve root: %v", err)
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
		iD, err := getSugId(DirSug, strings.TrimSuffix(name, " "))
		if err != nil {
			fmt.Printf(string(colorRed), "file or dir not exist: "+err.Error())
			glog.Errorln("file or dir not exist: " + err.Error())
			// return nil, err
		}
		id = iD
	}
	item, err := utils.StartSrv(drive.DriveScope).
		Files.Get(id).
		// Files.Get("root").
		Fields("id, name, mimeType, parents, owners, createdTime").
		Do()
	if err != nil {
		println("shit happened: ", err.Error())
		glog.Fatalf("Unable to retrieve root: %v", err)
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
		DirSug = dirInfo(s2)
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
	iD, err := getSugId(AllSug, strings.TrimSuffix(fil[0], " "))
	if err != nil {
		glog.Errorln("file or dir not exist: " + err.Error())
		return err
	}
	file, err := utils.StartSrv(drive.DriveScope).Files.Get(iD).Fields("id, name, mimeType, parents, createdTime").Do()
	if err != nil {
		glog.Errorln("file or dir not exist: ", err.Error())
		return err
	}

	if file.Id == ii.RootId {
		return errors.New("The root folder should not be moved")
	}

	if len(breakDown(fil[1])) > 1 { // move to another folder
		newParentName := strings.Trim(breakDown(fil[1])[len(breakDown(fil[1]))-2], " ") // move to another folder
		newName := strings.Trim(breakDown(fil[1])[len(breakDown(fil[1]))-1], " ")       // change item name
		iD, err := getSugId(AllSug, newParentName)
		if err != nil {
			glog.Errorln("file or dir not exist: ", err.Error())
			return err
		}

		var parents string
		if len(file.Parents) > 0 {
			parents = file.Parents[0]
		}
		newFile, err := utils.StartSrv(drive.DriveScope).Files.Update(file.Id, &drive.File{
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
		newFile, err := utils.StartSrv(drive.DriveScope).Files.Update(file.Id, &drive.File{
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
	// msg("Downloading... "+ humanize.Bytes(wc.Total)+ " complete ")
}

// getSugDec ...
func GetSugDec(sug *[]prompt.Suggest, text string) string {

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

// downld ...
func downld(id, target string) error {
	glog.V(8).Info("this is download debug ", id, target)
	// drive.DriveReadonlyScope
	fgc := utils.StartSrv(drive.DriveScope).Files.Get(id)
	fgc.Header().Add("alt", "media")
	resp, err := fgc.Download()

	glog.V(8).Info("this is download x0")
	if err != nil {
		glog.V(8).Info("this is download x0", err.Error())
		glog.Fatalf("Unable to retrieve files: %v", err)
		return err
	}
	glog.V(8).Info("this is download x1", id)
	defer resp.Body.Close()
	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	fileName := strings.Trim(target, " ") + "/" + GetSugDec(FileSug, id)
	glog.V(8).Info("this is download x1.1 ", fileName)
	out, err := os.Create(fileName + ".tmp")
	if err != nil {
		return err
	}
	glog.V(8).Info("this is download x2 ", fileName)
	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
	}
	glog.V(8).Info("this is download x3")
	// The progress use the same line so print a new line once it's finished downloading
	fmt.Print("\n")

	// Close the file without defer so it can happen before Rename()
	out.Close()

	// println("this is download x3-1")
	if err = os.Rename(fileName+".tmp", fileName); err != nil {
		// println("this is download x3-2", err.Error())
		// return err
		glog.Fatalf("Unable to save files: %v", err)
	}
	// println("this is download x4")
	return nil
}

// download file
func Download(cmd string) error {
	//TODO: download file

	//1 transfer file name to id
	var id string
	if !strings.Contains(cmd, ">") {
		fmt.Printf(string(colorRed), "Wrong path format, please use \"h\" get help")
		return errors.New("Wrong path format, please use \"h\" get help")
	}
	fil := strings.Split(strings.Split(cmd, "d ")[1], ">")
	iD, err := getSugId(AllSug, strings.TrimSuffix(fil[0], " "))
	if err != nil {
		fmt.Printf(string(colorRed), "file or dir not exist: "+err.Error())
		glog.Errorln("file or dir not exist: " + err.Error())
		return err
	}
	id = iD
	//2 check the id whether is file or folder
	// var fis, fods []string
	files, folders, err := utils.GetFilesAndFolders(id, strings.Trim(fil[0], " "))
	if err != nil {
		glog.Errorln("Download failed: ", err.Error())
		return err
	}
	for key, value := range files {
			glog.V(6).Info("download files: ",key," : ", value)

	}

	for _, value := range folders {
		glog.V(6).Info("download folders: ", value)

	}
	utils.ClearDownloadMap(files)
	//3 start download
	// err = downld(id, fil[1])
	// if err != nil {
	// 	glog.Errorln("Download failed: ", err.Error())
	// 	return err
	// }
	return nil
}

// Downloadd file by id
func Downloadd(cmds []string) error {
	//TODO: download file by id
	//1 transfer file name to id
	//2 check the id whether is file or folder
	//3 start download

	err := downld(cmds[1], cmds[2])
	if err != nil {
		glog.Errorln("Download failed: ", err.Error())
		return err
	}
	return nil
}

