package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	// "math/rand"
	// "flag"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	// "io/ioutil"
	// "sync"
	// "strconv"
	"github.com/golang/glog"

	// "os/exec"
	// "encoding/json"
	"github.com/c-bata/go-prompt"
	"github.com/cheynewallace/tabby"
	"github.com/dustin/go-humanize"
	. "github.com/logrusorgru/aurora"
	"github.com/roleyzhang/GoClue/utils"
	"github.com/theckman/yacspin"
	"go.uber.org/ratelimit"
	"golang.org/x/net/context"
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
var IdAllSug *[]prompt.Suggest

var TypesSug *[]prompt.Suggest
var RoleSug *[]prompt.Suggest
var GmailSug *[]prompt.Suggest
var DomainSug *[]prompt.Suggest

var cfg *yacspin.Config
var pthSep string

// var colorGreen string
// var colorCyan string
var colorYellow string
var colorRed string

var commands map[string]string

var Ps PromptStyle

var Ii ItemInfo

var perPageSize int64

type ItemInfo struct {
	// item       *drive.File
	Path      map[string]string
	RootId    string
	ItemId    string
	maxLength int
}

type PromptStyle struct {
	Pre      string
	Gap      string
	FolderId string
	Info     string
	Status   string
}

func init() {
	// colorGreen = "\033[32m%26s  %s\t%s\t%s\t%s\n"
	// colorCyan = "\033[36m%26s  %s\t%s\t%s\t%s\n"
	colorYellow = "\033[33m%s %s %s\n"
	colorRed = "\033[31m%s\n"

	tps := []prompt.Suggest{
		{Text: "user", Description: "identifies the scope of user"},
		{Text: "group", Description: "identifies the scope of group"},
		{Text: "domain", Description: "identifies the scope of domain"},
		{Text: "anyone", Description: "identifies the scope of anyone"},
	}
	TypesSug = &tps
	roles := []prompt.Suggest{
		{Text: "organizer", Description: "defines what users can do with a file or folder"},
		{Text: "fileOrganizer", Description: "defines what users can do with a file or folder"},
		{Text: "writer", Description: "defines what users can do with a file or folder"},
		{Text: "commenter", Description: "defines what users can do with a file or folder"},
		{Text: "reader", Description: "defines what users can do with a file or folder"},
	}
	RoleSug = &roles

	GmailSug = utils.LoadproSugg("mail.json")
	DomainSug = utils.LoadproSugg("domain.json")
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

	Ii = ItemInfo{
		Path:      make(map[string]string),
		RootId:    "",
		ItemId:    "",
		maxLength: 40,
	}

	Ps = PromptStyle{
		Pre:      "$0[$1 $2]",
		Gap:      ">>>",
		FolderId: "",
		Info:     "",
		Status:   "",
	}

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
	pthSep = string(os.PathSeparator)
	perPageSize = 40
}

// getSugId ...
func (ii *ItemInfo) getSugId(sug *[]prompt.Suggest, text string) (string, error) {
	// fmt.Println(text)
	if sug != nil {
		for _, v := range *sug {
			if v.Text == text {
				return v.Description, nil
			}
		}
	}
	qString := "name='" + text + "'" + " and '" + ii.ItemId + "' in parents " + " and trashed=false"

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

//-----------------------------
// type Callback func(msg string)

// func SetPrefix(msgs string, ii *ItemInfo, callback Callback ) {
// 	// glog.V(8).Info("SetPrefix: ",msgs, ii.ItemId, len(*DirSug) )
// 	folderId := ii.ItemId
// 	if DirSug != nil {
// 		folderName := GetSugDec(DirSug, folderId)
// 		callback(folderName + msgs)

// 	}
// }

func (ps *PromptStyle) SetPrefix(msgs string) {
	wc := 0
	switch wc {
	case 0:
		done := make(chan struct{})
		go func(mas string) {
			ps.Info = msgs
			done <- struct{}{}
		}(msgs)
		<-done
	case 1:
		done := make(chan struct{})
		go func(mas string) {
			ps.Status = msgs
			done <- struct{}{}
		}(msgs)
		<-done
	}
	// go func(mas string) {
	// 	ps.Info = msgs
	// }(msgs)
}

// func (ps *PromptStyle) SetStatus(msgs string) {
// 	// Create a channel to push an empty struct to once we're done
// 	done := make(chan struct{})
// 	go func(mas string) {
// 		ps.Status = msgs
// 		// Push an empty struct once we're done
// 		done <- struct{}{}
// 	}(msgs)
// 	<-done
// }

func (ps *PromptStyle) SetDynamicPrefix() (string, bool) {
	// glog.V(8).Info("SetPrefix: ",msgs, ii.ItemId, len(*DirSug) )
	var result string
	if DirSug != nil {
		folderName := GetSugDec(DirSug, ps.FolderId)
		value := ps.Pre
		r := strings.NewReplacer(
			"$0", ps.Status,
			"$1", ps.Info,
			"$2", folderName,
		)
		result = r.Replace(value)
		return (result + ps.Gap), true
	}
	return (result + ps.Gap), true
}

// getRoot ...
func (ps *PromptStyle) GetRoot(ii *ItemInfo) {
	dirInfo := utils.GetSugInfo()
	item, err := utils.StartSrv(drive.DriveScope).
		// Files.Get(id).
		Files.Get("root").
		Fields("id, name, mimeType").
		Do()
	if err != nil {
		// glog.("shit happened: ", err.Error())
		glog.Errorf("Unable to retrieve root: %v", err)
		// return nil
	}
	if item.MimeType == "application/vnd.google-apps.folder" {
		ii.Path[item.Id] = item.Name
		ii.ItemId = item.Id
		ii.RootId = item.Id
		ps.FolderId = item.Id
		// setting the prompt.Suggest
		s2 := prompt.Suggest{Text: item.Name, Description: item.Id}
		DirSug = dirInfo(s2)
	}
}

// // msg ...
// func msg(message string) {
// 	// glog.V(8).Info("msg: ", message)
// 	LivePrefixState.LivePrefix = message + ">>> "
// 	LivePrefixState.IsEnable = true
// }
//-------------------------------
// breakDown ...
func breakDown(path string) []string {
	return strings.Split(path, "/")
}

// print the request result
func (ii *ItemInfo) ShowResult(
	page map[int]string,
	counter int,
	param, cmd, scope string) *drive.FileList { // This should testing by change the authorize token
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

	//-----yacspin-----------------
	spinner, err := yacspin.New(*cfg)
	if err != nil {
		glog.V(8).Info("Spin run error", err.Error())
	}
	err = spinner.CharSet(yacspin.CharSets[31])
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
	dirInfo := utils.GetSugInfo()
	fileInfo := utils.GetSugInfo()
	allInfo := utils.GetSugInfo()
	idfileInfo := utils.GetSugInfo()
	iddirInfo := utils.GetSugInfo()
	idAllInfo := utils.GetSugInfo()

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
		iD, err := ii.getSugId(AllSug, cmd)
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
		Fields("nextPageToken, files(id, name, mimeType, owners, parents, createdTime, permissions, sharingUser)").
		PageToken(page[counter]).
		// OrderBy("modifiedTime").
		Do()

	if err != nil {
		fmt.Printf(string(colorRed), "Unable to retrieve files: %v", err.Error())
		return nil
		// uncomment below will cause 500 error and program exit why?
		// glog.Errorf("Unable to retrieve files: %v", err)
	}
	fmt.Println("Files:")
	if len(r.Files) == 0 {
		fmt.Printf(string(colorYellow), "No files found.", "", "")
		// fmt.Println("No files found.")
	} else {
		t := tabby.New()
		t.AddHeader("NAME", "ID", "TYPE", "OWNER", "CREATED TIME")
		for _, i := range r.Files {
			if i.MimeType == "application/vnd.google-apps.folder" {
				// for _, value := range i.Permissions{
				// fmt.Println(i.Name, i.Id, i.MimeType,len(i.Permissions), value.EmailAddress, value.Id, value.Role)
				// fmt.Println(i.Name, i.Id, i.MimeType,len(i.Permissions), i.SharingUser)
				// }
				var name string
				var types string
				if len(i.Name) > ii.maxLength {
					name = i.Name[:ii.maxLength] + "..."
				} else {
					name = i.Name
				}
				if len(i.Permissions) > 1 {
					types = fmt.Sprintf("%s %s", Brown("*S*"), strings.Split(i.MimeType, "/")[1])
				} else {
					types = strings.Split(i.MimeType, "/")[1]
				}
				t.AddLine(Bold(Cyan(name)),
					Bold(Cyan(i.Id)),
					Bold(Cyan(types)),
					Bold(Cyan(i.Owners[0].DisplayName)),
					Bold(Cyan(i.CreatedTime)))
				// fmt.Printf(string(colorGreen), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.CreatedTime)
				s := prompt.Suggest{Text: i.Id, Description: i.Name}
				s2 := prompt.Suggest{Text: i.Name, Description: i.Id}
				DirSug = dirInfo(s2)
				AllSug = allInfo(s2)
				IddirSug = iddirInfo(s)
				IdAllSug = idAllInfo(s)
				// 	s := prompt.Suggest{Text: i.Id, Description: i.Name}
				// 	dirSug = dirInfo(s)
			} else {
				// fmt.Printf(string(colorCyan), i.Name, i.Id, i.MimeType, i.Owners[0].DisplayName, i.Parents, i.CreatedTime)
				// for _, value := range i.Permissions{
				// 	fmt.Println(i.Name, i.Id, i.MimeType,len(i.Permissions), value.EmailAddress,value.Id, value.Role)
				// fmt.Println(i.Name, i.Id, i.MimeType,len(i.Permissions), i.SharingUser)
				// }
				var name string
				var types string
				if len(i.Name) > ii.maxLength {
					name = i.Name[:ii.maxLength] + "..."
				} else {
					name = i.Name
				}
				if len(i.Permissions) > 1 {
					types = fmt.Sprintf("%s %s", Brown("*S*"), strings.Split(i.MimeType, "/")[1])
				} else {
					types = strings.Split(i.MimeType, "/")[1]
				}
				if i.MimeType == "application/vnd.google-apps.shortcut" {
					t.AddLine(Gray(name),
						Gray(i.Id),
						Gray(types),
						Gray(i.Owners[0].DisplayName),
						Gray(i.CreatedTime))

				} else {
					t.AddLine(Green(name),
						Green(i.Id),
						Green(types),
						Green(i.Owners[0].DisplayName),
						Green(i.CreatedTime))
				}
				s := prompt.Suggest{Text: i.Id, Description: i.Name}
				s2 := prompt.Suggest{Text: i.Name, Description: i.Id}
				FileSug = fileInfo(s2)
				AllSug = allInfo(s2)
				IdfileSug = idfileInfo(s)
				IdAllSug = idAllInfo(s)
			}
		}
		t.Print()
	}
	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
	return r
}

//  generate folder path...
func PathGenerate(path, level string) {

	pathInfo := utils.GetSugInfo()
	// home, _ :=os.UserHomeDir()
	if path == "HOME" {
		// only list folders
		// cmd := exec.Command("tree", "-f", "-i", "-d", os.Getenv(path))

		cmd := exec.Command("tree", "-f", "-L", level, "-i", "-d", os.Getenv(path))
		output, _ := cmd.Output()
		for _, m := range strings.Split(string(output), "\n") {
			// fmt.Printf("metric: %s\n", m)
			s := prompt.Suggest{Text: m, Description: ""}
			PathSug = pathInfo(s)
		}

		// files := make([]string, 0)
		// GetLocalItems(os.Getenv(path), true, &files)
		// for _, file := range files {
		// 	// fmt.Println(string)
		// 	s := prompt.Suggest{Text: strings.Replace(file,home+pthSep,"",1), Description: ""}
		// 	// s := prompt.Suggest{Text: file, Description: ""}
		// 	PathSug = pathInfo(s)
		// }
	} else {
		// list files & folder
		// cmd := exec.Command("tree", "-f", "-i", path)

		cmd := exec.Command("tree", "-f", "-L", level, "-i", path)
		output, _ := cmd.Output()
		for _, m := range strings.Split(string(output), "\n") {
			// fmt.Printf("metric: %s\n", m)
			s := prompt.Suggest{Text: m, Description: ""}
			PathSug = pathInfo(s)
		}

		// files := make([]string, 0)
		// GetLocalItems(os.Getenv(path), false, &files)
		// for _, file := range files {
		// 	// fmt.Println(file)
		// 	s := prompt.Suggest{Text: strings.Replace(file,home+pthSep,"",1), Description: ""}
		// 	PathSug = pathInfo(s)
		// }
	}

}

//  generate folder and file path...
func PathFileGenerate(path, level string) {
	pathInfo := utils.GetSugInfo()
	// home, _ :=os.UserHomeDir()

	if path == "HOME" {
		// only list folders
		// cmd := exec.Command("tree", "-f", "-i", "-d", os.Getenv(path))

		cmd := exec.Command("tree", "-f", "-L", level, "-i", os.Getenv(path))
		output, _ := cmd.Output()
		for _, m := range strings.Split(string(output), "\n") {
			// fmt.Printf("metric: %s\n", m)
			s := prompt.Suggest{Text: m, Description: ""}
			AllSug = pathInfo(s)
		}

		// files := make([]string, 0)
		// GetLocalItems(os.Getenv(path), true, &files)
		// for _, file := range files {
		// 	// fmt.Println(string)
		// 	s := prompt.Suggest{Text: strings.Replace(file,home+pthSep,"",1), Description: ""}
		// 	// s := prompt.Suggest{Text: file, Description: ""}
		// 	PathSug = pathInfo(s)
		// }
	} else {
		// list files & folder
		// cmd := exec.Command("tree", "-f", "-i", path)

		cmd := exec.Command("tree", "-f", "-L", level, "-i", path)
		output, _ := cmd.Output()
		for _, m := range strings.Split(string(output), "\n") {
			// fmt.Printf("metric: %s\n", m)
			s := prompt.Suggest{Text: m, Description: ""}
			AllSug = pathInfo(s)
		}

		// files := make([]string, 0)
		// GetLocalItems(os.Getenv(path), false, &files)
		// for _, file := range files {
		// 	// fmt.Println(file)
		// 	s := prompt.Suggest{Text: strings.Replace(file,home+pthSep,"",1), Description: ""}
		// 	PathSug = pathInfo(s)
		// }
	}
}

// SetPrefix ...

// rmd ... delete file by id
func (ii *ItemInfo) Rmd(id, types string) error {
	//TODO: delete file
	//-----yacspin-----------------
	spinner, err := yacspin.New(*cfg)
	if err != nil {
		glog.Error("Spin run error", err.Error())
	}
	if err := spinner.Frequency(100 * time.Millisecond); err != nil {
		glog.Error("Spin run error", err.Error())
	}
	// msg := fmt.Sprintf("   Getting %s information from server", id)
	// spinner.Suffix(msg)
	msgs := fmt.Sprintf("   Delete %s... %s ", id, Brown("done"))
	spinner.StopMessage(msgs)
	err = spinner.CharSet(yacspin.CharSets[9])
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
	file, err := utils.StartSrv(drive.DriveScope).Files.Get(id).Do()
	if err != nil {
		spinner.StopFailMessage("   file or dir not exist")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("file or dir not exist", err)
		}
		glog.Error("file or dir not exist: ", err.Error())
		return err
	}

	if id == ii.RootId {
		spinner.StopFailMessage("   The root folder should not be deleted")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("The root folder should not be deleted", err)
		}
		return errors.New("The root folder should not be deleted")
	}

	if types == "-r" && file.MimeType != "application/vnd.google-apps.folder" {
		spinner.StopFailMessage("   The delete item: item is not folder")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("The delete item: item is not folder", err)
		}
		return errors.New("The delete item: item is not folder")
	}

	err = utils.StartSrv(drive.DriveScope).Files.Delete(id).Do()

	if err != nil {
		spinner.StopFailMessage("   File or dir delete failed")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("File or dir delete failed", err)
		}
		glog.Errorln("file or dir delete failed: " + err.Error())
		return err
	}
	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
	return nil
}

// rm ... delete file
func (ii *ItemInfo) Rm(name, types string) error {
	//-----yacspin-----------------
	spinner, err := yacspin.New(*cfg)
	if err != nil {
		glog.Error("Spin run error", err.Error())
	}
	if err := spinner.Frequency(100 * time.Millisecond); err != nil {
		glog.Error("Spin run error", err.Error())
	}
	// msg := fmt.Sprintf("   Getting %s information from server", id)
	// spinner.Suffix(msg)
	msgs := fmt.Sprintf("   Delete %s... %s ", name, Brown("done"))
	spinner.StopMessage(msgs)
	err = spinner.CharSet(yacspin.CharSets[9])
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
	//TODO: delete file
	var id string
	iD, err := ii.getSugId(DirSug, strings.TrimSuffix(name, " "))
	if err != nil {
		// fmt.Printf(string(colorRed), "file or dir not exist: "+err.Error())
		glog.Errorln("file or dir not exist: ", err.Error())
		spinner.StopFailMessage("   file or dir not exist, or try to use rmd delete by id")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("file or dir not exist", err)
		}
		return err
	}
	id = iD

	file, err := utils.StartSrv(drive.DriveScope).Files.Get(id).Do()
	if err != nil {
		glog.Errorln("file or dir not exist: ", err.Error())
		spinner.StopFailMessage("   file or dir not exist")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("file or dir not exist", err)
		}
		return err
	}

	if id == ii.RootId {
		spinner.StopFailMessage("   The root folder should not be deleted")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("The root folder should not be deleted", err)
		}
		return errors.New("The root folder should not be deleted")
	}

	if types == "-r" && file.MimeType != "application/vnd.google-apps.folder" {
		spinner.StopFailMessage("   The delete item: item is not folder")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("The delete item: item is not folder", err)
		}
		return errors.New("The delete item: item is not folder")
	}

	err = utils.StartSrv(drive.DriveScope).Files.Delete(id).Do()

	if err != nil {
		spinner.StopFailMessage("   File or dir delete failed")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("File or dir delete failed", err)
		}
		glog.Errorln("File or dir delete failed: ", err.Error())
		return err
	}
	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
	return nil
}

// trash ...
func (ii *ItemInfo) Trash(name, types string) error {
	//TODO: trash file
	var id string
	iD, err := ii.getSugId(DirSug, strings.TrimSuffix(name, " "))
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

func (ii *ItemInfo) upload(file, parentId string) (*drive.File, error) {
	//-----yacspin-----------------
	spinner, err := yacspin.New(*cfg)
	if err != nil {
		glog.Error("Spin run error", err.Error())
	}
	if err := spinner.Frequency(100 * time.Millisecond); err != nil {
		glog.Error("Spin run error", err.Error())
	}
	msg := fmt.Sprintf("   Uploading %s to cloud ", file)
	spinner.Suffix(msg)
	msgs := fmt.Sprintf("... %s ", Brown("done"))
	spinner.StopMessage(msgs)
	err = spinner.CharSet(yacspin.CharSets[71])
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
	if parentId == "" {
		parentId = ii.ItemId
	}
	// fil := strings.Split(file, "u ")
	fi, err := os.Open(file)
	if err != nil {
		spinner.StopFailMessage("   Open local file failed")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("Try to open local file failed", err)
		}
		return nil, err
	}
	defer fi.Close()

	fileInfo, err := fi.Stat()
	if err != nil {
		spinner.StopFailMessage("   Get file stat failed")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("Try to get file stat failed", err)
		}
		return nil, err
	}
	u := &drive.File{
		Name:    filepath.Base(fileInfo.Name()),
		Parents: []string{parentId},
	}
	// ufile, err := startSrv(drive.DriveScope).Files.Create(u).Media(fi).Do()
	ufile, err := utils.StartSrv(drive.DriveScope).Files.
		Create(u).
		ResumableMedia(context.Background(), fi, fileInfo.Size(), "").
		ProgressUpdater(func(now, size int64) {
			// fmt.Printf("%d, %d\r", now, size)
			rate := now * 100 / size
			mesg := fmt.Sprintf("%d%%  %s / %s",
				Brown(rate),
				Brown(humanize.Bytes(uint64(now))),
				Brown(humanize.Bytes(uint64(size))))
			spinner.Message(mesg)
		}).
		Do()
	if err != nil {
		spinner.StopFailMessage("   Upload file failed")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("Try to upload file failed", err)
		}
		return nil, err
	}

	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
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

// createInDir...
func CreateInDir() func(name, path, parentId string) (map[string]string, *drive.File, error) {
	var parent string
	parents := make(map[string]string)
	return func(name, path, parentId string) (map[string]string, *drive.File, error) {
		parDir, parName := filepath.Split(path[0 : len(path)-1])
		glog.V(8).Info("fil: ", name, " parDir,", parDir, " , parName ", parName)
		parent = parents[parName]
		if parent == "" {
			parent = parentId
		}
		d := &drive.File{
			Name:     name,
			MimeType: "application/vnd.google-apps.folder",
			Parents:  []string{parent},
		}
		dir, err := utils.StartSrv(drive.DriveScope).Files.Create(d).Do()

		if err != nil {
			glog.Errorln("Could not create dir: ", err.Error())
			return nil, nil, err
		}

		fmt.Printf(string(colorYellow), dir.Name, "", " has been created")
		parents[dir.Name] = dir.Id
		return parents, dir, nil
	}
}

// // getRoot ...
// func (ii *ItemInfo) GetRoot() {
// 	dirInfo := getSugInfo()
// 	item, err := utils.StartSrv(drive.DriveScope).
// 		// Files.Get(id).
// 		Files.Get("root").
// 		Fields("id, name, mimeType, parents, owners, createdTime").
// 		Do()
// 	if err != nil {
// 		println("shit happened: ", err.Error())
// 		glog.Errorf("Unable to retrieve root: %v", err)
// 		// return nil
// 	}
// 	if item.MimeType == "application/vnd.google-apps.folder" {
// 		ii.Path[item.Id] = item.Name
// 		ii.ItemId = item.Id
// 		ii.RootId = item.Id
// 		// setting the prompt.Suggest
// 		s2 := prompt.Suggest{Text: item.Name, Description: item.Id}
// 		DirSug = dirInfo(s2)
// 	}
// }

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
		glog.Errorf("Unable to retrieve root: %v", err)
		// return nil
	}
	if item.MimeType == "application/vnd.google-apps.folder" || item.MimeType == "application/vnd.google-apps.shortcut" {
		ii.Path[item.Id] = item.Name
		ii.ItemId = item.Id
		if id == "root" {
			ii.RootId = item.Id
		}
		Ps.FolderId = item.Id
	}
}

// getNode ...
func (ii *ItemInfo) GetNode(cmd string) {
	// println(id)
	var id string

	if len(strings.Split(cmd, "cd ")) <= 1 {
		return
	}
	name := strings.Trim(strings.Split(cmd, "cd ")[1], " ")
	dirInfo := utils.GetSugInfo()

	if name == "root" || name == "My Drive" {
		id = "root"
	} else {
		iD, err := ii.getSugId(DirSug, strings.TrimSuffix(name, " "))
		if err != nil {
			fmt.Printf(string(colorRed), "file or dir not exist: "+err.Error())
			glog.Errorln("file or dir not exist: " + err.Error())
			return
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
		glog.Errorf("Unable to retrieve root: %v", err)
		// return nil
	}
	if item.MimeType == "application/vnd.google-apps.folder" || item.MimeType == "application/vnd.google-apps.shortcut" {
		ii.Path[item.Id] = item.Name
		ii.ItemId = item.Id
		if id == "root" {
			ii.RootId = item.Id
		}
		Ps.FolderId = item.Id

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
	iD, err := ii.getSugId(AllSug, strings.TrimSuffix(fil[0], " "))
	if err != nil {
		glog.Errorln("file or dir not exist: " + err.Error())
		return err
	}
	file, err := utils.StartSrv(drive.DriveScope).
		Files.Get(iD).Fields("id, name, mimeType, parents, createdTime").Do()
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
		iD, err := ii.getSugId(AllSug, newParentName)
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
	Total   uint64
	Spinner *yacspin.Spinner
	Amount  uint64
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
	// fmt.Printf("%s", strings.Repeat(" ", 35))
	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	// fmt.Printf(". %s complete", humanize.Bytes(wc.Total))

	rate := wc.Total * 100 / wc.Amount
	mesg := fmt.Sprintf("downloading  %d%%  %s / %s",
		Brown(rate), Brown(humanize.Bytes(wc.Total)), Brown(humanize.Bytes(wc.Amount)))
	wc.Spinner.Message(mesg)
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
func downld(id, target, filename, mimeType, path string) error {
	glog.V(8).Info("this is download debug ", id, target)
	//-----yacspin-----------------
	spinner, err := yacspin.New(*cfg)
	if err != nil {
		glog.Error("Spin run error", err.Error())
	}
	if err := spinner.Frequency(300 * time.Millisecond); err != nil {
		glog.Error("Spin run error", err.Error())
	}
	msg := fmt.Sprintf("Downloading %s", target)
	spinner.Suffix(msg)
	msgs := fmt.Sprintf("...to %s %s", path, Brown("done"))
	spinner.StopMessage(msgs)
	err = spinner.CharSet(yacspin.CharSets[50])
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
	var resp *http.Response
	var fileName string
	// Download binary file
	if mimeType != "application/vnd.google-apps.document" &&
		mimeType != "application/vnd.google-apps.spreadsheet" &&
		mimeType != "application/vnd.google-apps.form" &&
		mimeType != "application/vnd.google-apps.drawing" &&
		mimeType != "application/vnd.google-apps.presentation" &&
		mimeType != "application/vnd.google-apps.script" {

		fgc := utils.StartSrv(drive.DriveScope).Files.Get(id)
		fgc.Header().Add("alt", "media")
		resp, err = fgc.Download()
		fileName = strings.Trim(target, " ")
	} else {
		glog.V(8).Info("this is download x0", mimeType)
		// Download Google docs file
		resp, err = utils.StartSrv(drive.DriveScope).Files.
			Export(id, "application/zip").Download()
		fileName = strings.Trim(target, " ") + ".zip"
	}
	if err != nil {
		glog.V(8).Info("this is download x0", err.Error())
		// glog.Errorf("Unable to retrieve files: %v", err)
		if err := spinner.StopFail(); err != nil {
			return err
		}
		return err
	}
	glog.V(8).Info("this is download x1", id)
	defer resp.Body.Close()
	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	// fileName := strings.Trim(target, " ") + "/" + GetSugDec(FileSug, id)
	glog.V(8).Info("this is download x1.1 ", fileName, path)
	// spinner.Message("Creating tmp file")

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			glog.Errorln("Create folder failed ", err.Error())
			return err
		}
	}
	out, err := os.Create(fileName + ".tmp")
	if err != nil {
		return err
	}
	glog.V(8).Info("this is download x2 ", fileName)
	// Create our progress reporter and pass it to be used alongside our writer
	//-----------------------------
	counter := &WriteCounter{Spinner: spinner, Amount: uint64(resp.ContentLength)}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
	}

	// if _, err = io.Copy(out, resp.Body); err != nil {
	// 	out.Close()
	// 	return err
	// }
	//-----------------------------
	glog.V(8).Info("this is download x3")
	// The progress use the same line so print a new line once it's finished downloading
	// fmt.Print("\n")

	// Close the file without defer so it can happen before Rename()
	out.Close()

	// println("this is download x3-1")
	if err = os.Rename(fileName+".tmp", fileName); err != nil {
		// println("this is download x3-2", err.Error())
		// return err
		// glog.Errorf("Unable to save files: %v", err)
		glog.Errorf("Unable to save files: %v", err)
	}
	if err := spinner.Stop(); err != nil {

		glog.Errorf("Spinner err: %v", err)
	}

	return nil
}

// download file by name
func (ii *ItemInfo) Download(cmd string) error {
	//TODO: download file

	//1 transfer file name to id
	var id string
	if !strings.Contains(cmd, ">") {
		fmt.Printf(string(colorRed), "Wrong path format, please use \"h\" get help")
		return errors.New("Wrong path format, please use \"h\" get help")
	}
	fil := strings.Split(strings.Split(cmd, "d ")[1], ">")
	iD, err := ii.getSugId(AllSug, strings.TrimSuffix(fil[0], " "))
	if err != nil {
		fmt.Printf(string(colorRed), "file or dir not exist: "+err.Error())
		glog.Errorln("file or dir not exist: " + err.Error())
		return err
	}
	id = iD
	//2 start download
	targetPath := strings.Trim(fil[1], " ") + string(os.PathSeparator) + strings.Trim(fil[0], " ")
	StartingDownload(id, targetPath)
	return nil
}

// Downloadd file by id
func Downloadd(cmds []string) error {
	//TODO: download file by id
	//1 transfer file name to id
	//2 check the id whether is file or folder
	//3 start download

	// err := downld(cmds[1], cmds[2], cmds[1])
	// if err != nil {
	// 	glog.Errorln("Download failed: ", err.Error())
	// 	return err
	// }
	if len(cmds) < 3 {
		return errors.New("command line lack param")
	}

	item, err := utils.StartSrv(drive.DriveScope).
		// Files.Get(id).
		Files.Get(cmds[1]).
		Fields("id, name, mimeType").
		Do()
	if err != nil {
		// glog.("shit happened: ", err.Error())
		// glog.Errorf("Unable to retrieve root: %v", err)
		return err
	}
	if item.MimeType == "application/vnd.google-apps.folder" {
		glog.V(6).Info("B7: ", cmds[2])
		StartingDownload(cmds[1], cmds[2]+string(os.PathSeparator)+item.Name)
	} else {
		glog.V(6).Info("B7: ", cmds[2])
		StartingDownload(cmds[1], cmds[2])
	}
	return nil
}

func GetAllDriveItems(id, pageToken string, files *[]*drive.File) {

	qString := "'" + id + "' in parents and trashed=false"
	r, err := utils.StartSrv(drive.DriveScope).Files.List().
		Q(qString).
		PageSize(perPageSize).
		Fields("nextPageToken, files(id, name, mimeType)").
		PageToken(pageToken).
		Do()
	if err != nil {
		glog.Errorln("file or dir not exist: ", err.Error())
	}
	*files = append(*files, r.Files...)
	glog.V(8).Info("files1 len:  ", len(*files))
	if int64(len(r.Files)) == perPageSize {
		pageToken := r.NextPageToken
		GetAllDriveItems(id, pageToken, files)
	}
	// return files
}

func generatorDownloader(id, path string, out, stop chan string) {
	// out := make(chan int, 50)
	go func() {
		//-----yacspin-----------------
		spinner, err := yacspin.New(*cfg)
		if err != nil {
			glog.Error("Spin run error", err.Error())
		}
		if err := spinner.Frequency(100 * time.Millisecond); err != nil {
			glog.Error("Spin run error", err.Error())
		}
		msg := fmt.Sprintf("   Getting %s information from server", id)
		spinner.Suffix(msg)
		spinner.StopCharacter("->")
		msgs := fmt.Sprintf("... %s ", Brown("done"))
		spinner.StopMessage(msgs)
		err = spinner.CharSet(yacspin.CharSets[9])
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
		// qString := "'" + id + "' in parents"
		// qString := "'" + id + "' in parents and trashed=false"
		// // glog.V(8).Info("B1: ", qString)
		// item, err := utils.StartSrv(drive.DriveScope).Files.List().
		// 	Q(qString).PageSize(perPageSize).
		// 	Fields("nextPageToken, files(id, name, mimeType)").
		// 	Do()
		// if err != nil {
		// 	glog.Errorln("file or dir not exist: ", err.Error())
		// 	return
		// }
		files := make([]*drive.File, 0)
		GetAllDriveItems(id, "", &files)
		// glog.V(8).Info("B2: ", item.Files, len(item.Files))
		if len(files) == 0 {
			fil, err := utils.StartSrv(drive.DriveScope).Files.Get(id).Do()
			if err != nil {
				glog.Error("file or dir not exist: ", err.Error())
			}
			switch fil.MimeType {
			case "application/vnd.google-apps.shortcut":
				spinner.StopFailMessage("   Google Drive Shortcut cannot be downloaded")
				if err := spinner.StopFail(); err != nil {
					glog.Error("B3 shortcut: ", err)
				}
				stop <- "stop"
			case "application/vnd.google-apps.folder":
				spinner.StopFailMessage("   Empty folder cannot be downloaded")
				if err := spinner.StopFail(); err != nil {
					glog.V(8).Info("Try to downloading empty folder", err)
				}
				stop <- "stop"
			default:
				//send id to out chan
				glog.V(8).Info("B3 default: ", path+pthSep+fil.Name, fil.MimeType, fil.ShortcutDetails)
				pat := path + "-/-" +
					fil.Id + "-/-" +
					fil.MimeType + "-/-" +
					strings.Split(path, fil.Name)[0]
				out <- pat
				glog.V(8).Info("B3.1 default: ", pat)
				// make a global spinner
			}

		}
		//speed limit of google drive api 1000 requests per 100 seconds
		rl := ratelimit.New(5) // per second 5 requests
		prev := time.Now()
		// strf = path
		var pat string
		var patt string
		for _, file := range files {
			// glog.V(8).Info("B4: ")
			now := rl.Take()
			if file.MimeType == "application/vnd.google-apps.folder" {
				patt = pthSep + file.Name
				// strd = strd + pthSep + file.Name
				// glog.V(8).Info("B5: ", path + patt)
				out <- path + patt
				go generatorDownloader(file.Id, path+patt, out, stop)
			} else {
				// pat = strd + pthSep + file.Name + "-/-" + file.Id + "-/-" + file.MimeType + "-/-" + path
				pat = path + patt + pthSep + file.Name + "-/-" +
					file.Id + "-/-" + file.MimeType + "-/-" + path + patt
				glog.V(8).Info("B6: ", pat)
				glog.V(8).Info("B6-1: ", patt)
				pat = strings.Replace(pat, patt, "", 1)
				glog.V(8).Info("B6-2: ", pat)
				// out <- path + pat
				out <- pat
			}
			now.Sub(prev)
			prev = now
		}
		if err := spinner.Stop(); err != nil {
			glog.Errorf("Spinner err: %v", err)
		}
	}()
}

func downloader(id int, c chan string) {
	for n := range c {
		time.Sleep(time.Second)
		// fmt.Printf("Downloader %d received %s\n",
		// 	id, n)
		if strings.Contains(n, "-/-") {
			target := strings.Split(n, "-/-")
			glog.V(8).Infoln("------n >: ", target[3])
			// glog.V(8).Infoln("Starting downloading...", target[0], " ", target[1])
			//3 start download
			err := downld(target[1], target[0], target[1], target[2], target[3])
			if err != nil {
				glog.Errorln("Download failed: ", err.Error())
			}
		} else {
			// glog.V(8).Infoln("Creating folder...", n)
			_, err := os.Stat(n)
			if os.IsNotExist(err) {
				err := os.MkdirAll(n, 0755)
				if err != nil {
					glog.Errorln("Create folder failed ", err.Error())
				}
			}
			// err := os.MkdirAll(n, os.ModePerm)
		}
	}
}

func createDownloader(id int) chan<- string {
	c := make(chan string)
	go downloader(id, c)
	return c
}

func StartingDownload(id, path string) {
	glog.V(6).Info("download files: ", id, " : ", path)
	out := make(chan string)
	stop := make(chan string)
	// var fd string
	// var fls string
	generatorDownloader(id, path, out, stop)
	var downloader = createDownloader(0)
	var i, j int
	var values []string
	// tm := time.After(69 * time.Second)
	tick := time.NewTicker(5 * time.Second)

	for {
		var activeDownloader chan<- string
		var activeValue string
		if len(values) > 0 {
			activeDownloader = downloader
			activeValue = values[0]
		}

		select {
		case n := <-out:
			//add task to buffer
			values = append(values, n)
			j++
		case activeDownloader <- activeValue:
			// do task
			values = values[1:]
			i++
		// case <-time.After(800 * time.Millisecond):
		// 	fmt.Println("timeout")
		case <-tick.C:
			// if buffer =0 and task count = full buffer size then jump out
			if len(values) == 0 && i == j {
				return
			}
			// case <-tm:
			// 	fmt.Println("bye")
			// return
		case <-stop:
			return
		}
	}
}

func visit(files *[]string, isDir bool) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		// glog.V(8).Info(info)
		if info != nil {
			// checking permission
			if m := info.Mode(); m&(1<<2) == 0 {
				// glog.V(8).Info(m)
				return nil
			}
			if err != nil {
				// glog.V(8).Info(info.Mode().Perm().String)
				glog.Error(err)
			}
			// if strings.HasPrefix(info.Name(), ".") {
			// 	// glog.V(8).Info(path)
			// 	return nil
			// }
			// if info.IsDir() && noDir {
			// 	return nil
			// }
			if !info.IsDir() && isDir {
				return nil
			}
			*files = append(*files, path)

		}
		return nil
	}
}

// path is local path, noDir is switch list folder
func GetLocalItems(path string, isDir bool, files *[]string) {
	// var files []string

	fil := strings.Split(path, "u ")
	// root := "/some/folder/to/scan"
	err := filepath.Walk(fil[1], visit(files, isDir))
	if err != nil {
		// panic(err)
		glog.Error(err)
	}
	// return files
}

// Upload function
func (ii *ItemInfo) UpLod(file, scope string) {
	//-----yacspin-----------------
	spinner, err := yacspin.New(*cfg)
	if err != nil {
		glog.Error("Spin run error", err.Error())
	}
	if err := spinner.Frequency(100 * time.Millisecond); err != nil {
		glog.Error("Spin run error", err.Error())
	}
	// msg := fmt.Sprintf("   Getting %s information from server", id)
	// spinner.Suffix(msg)
	msgs := fmt.Sprintf("... %s ", Brown("done"))
	spinner.StopMessage(msgs)
	err = spinner.CharSet(yacspin.CharSets[29])
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
	// var tmpId string
	var parns map[string]string
	createInDir := CreateInDir()

	files := make([]string, 0)
	GetLocalItems(file, false, &files)
	for _, file := range files {
		// home, _ := os.UserHomeDir()
		// fmt.Println("FILE: ",file, "  ",strings.Replace(file, home+string(os.PathSeparator), "", 1),  "===== ", len(files))
		dir, fil := filepath.Split(file)
		// glog.V(8).Info("dir: ", dir, "   file: ", fil, " ----: ", utils.IsDir(file))

		if utils.IsDir(file) {
			// glog.V(8).Info(ii.ItemId)
			qString := "name ='" + fil +
				"' and mimeType = 'application/vnd.google-apps.folder' and '" +
				ii.ItemId + "' in parents and trashed = false"
			// glog.V(5).Infoln("qString: ", qString)
			r, err := utils.StartSrv(scope).Files.List().
				Q(qString).
				PageSize(10).
				Fields("nextPageToken, files(id, name, mimeType, owners, parents, createdTime)").
				// OrderBy("modifiedTime").
				Do()
			if err != nil {
				fmt.Printf(string(colorRed), "Unable to retrieve files: %v", err.Error())
				// uncomment below will cause 500 error and program exit why?
				glog.Errorf("Unable to retrieve files: %v", err)
			}
			if len(r.Files) != 0 {
				spinner.StopFailMessage("   Cloud already has the folder")
				if err := spinner.StopFail(); err != nil {
					glog.V(8).Info("Try to upload folder failed", err)
				}
				return
			} else {
				parents, _, err := createInDir(fil, dir, ii.ItemId)
				if err != nil {
					glog.Error(err)
					return
				}
				// tmpId = dr.Id
				parns = parents
			}
		} else {
			if spinner.Active() {
				if err := spinner.Pause(); err != nil {
					glog.Error(err)
				}
			}
			_, parName := filepath.Split(dir[0 : len(dir)-1])
			// glog.V(8).Info("file: ",file," id: ", tmpId, " parensID: ", parns[parName], " parDir: ", parDir)
			_, err := ii.upload(file, parns[parName])
			if err != nil {
				spinner.StopFailMessage("   File upload failed")
				if err := spinner.StopFail(); err != nil {
					glog.V(8).Info("Try to upload file failed", err)
				}
				return
			}
			if !spinner.Active() {
				if err := spinner.Unpause(); err != nil {
					glog.Error(err)
				}
			}
		}
	}
	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
}

//------------------------------TESTING BELOW
func (ii *ItemInfo) Share(idorName, types, role, gmail, domain string, isByName bool) {
	// If the type is user or group, provide an emailAddress. If the type is domain, provide a domain.
	// Type : user, group, domain, anyone - all are small letter
	// *EmailAddress : use any google account: email or groups address
	// *Role : organizer/owner	fileOrganizer	writer	commenter	reader
	// Domain : use any Google domain address

	//-----yacspin-----------------
	spinner, err := yacspin.New(*cfg)
	if err != nil {
		glog.Error("Spin run error", err.Error())
	}
	if err := spinner.Frequency(100 * time.Millisecond); err != nil {
		glog.Error("Spin run error", err.Error())
	}
	// msg := fmt.Sprintf("   Getting %s information from server", id)
	// spinner.Suffix(msg)
	msgs := fmt.Sprintf("...Share %s %s ", idorName, Brown("done"))
	spinner.StopMessage(msgs)
	err = spinner.CharSet(yacspin.CharSets[29])
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
	//-----name to id--------------
	var id string
	if isByName {
		iD, err := ii.getSugId(DirSug, strings.TrimSuffix(idorName, " "))
		if err != nil {
			spinner.StopFailMessage("   File or dir not exist, maybe file/folder name include space, try shared command by file/folder ID")
			if err := spinner.StopFail(); err != nil {
				glog.V(8).Info(" File or dir not exist", err)
			}
			glog.Errorln("file or dir not exist: ", err.Error())
			return
		}
		id = iD
	} else {
		id = idorName
	}
	//-----name to id--------------
	var permisn *drive.Permission = new(drive.Permission)
	permisn.EmailAddress = gmail
	permisn.Role = role
	permisn.Type = types
	permisn.Domain = domain

	_, errs := utils.StartSrv(drive.DriveScope).Permissions.Create(id, permisn).Do()

	if errs != nil {
		spinner.StopFailMessage("   Unable to create share item")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info(" Unable to create share item", err)
		}
		glog.Errorf("Unable to create share item: %v", err)
	}

	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}

	if GmailSug != nil {
		if !utils.IsContain(*GmailSug, gmail) {
			*GmailSug = append(*GmailSug, prompt.Suggest{Text: gmail, Description: ""})
		}
	}

	if DomainSug != nil {
		if !utils.IsContain(*DomainSug, domain) {
			*DomainSug = append(*DomainSug, prompt.Suggest{Text: domain, Description: ""})
		}
	}
}

func Lo() {
	// DomainSug = domainInfo(prompt.Suggest{Text: "www.roleyzhang.com", Description: ""})
	glog.V(8).Info(utils.GetAppHome())

	//----Load mail & domain Sug
	//----mail part
	//----domain part
	// domainInfo := getSugInfo()
	// domain, _ := ioutil.ReadFile(utils.GetAppHome() + string(os.PathSeparator) + "domain.json")
	// ddata := []prompt.Suggest{}
	// _ = json.Unmarshal([]byte(domain), &ddata)

	// for _, value := range ddata {
	// 	p := prompt.Suggest{Text: value.Text, Description: value.Description}
	// 	DomainSug = domainInfo(p)
	// 	// fmt.Println("Product Id: ", value.Text)
	// 	// fmt.Println("Quantity: ", value.Description)
	// }

	// file, _ := json.MarshalIndent(GmailSug, "", " ")
	// _ = ioutil.WriteFile(utils.GetAppHome()+pthSep+"mail.json", file, 0644)

	// file2, _ := json.MarshalIndent(DomainSug, "", " ")
	// _ = ioutil.WriteFile(utils.GetAppHome()+pthSep+"doman.json", file2, 0644)
	// var permisn *drive.Permission = new(drive.Permission)
	// permisn.EmailAddress = ""//"zhangroley@gmail.com"
	// permisn.Role = "writer"
	// permisn.Type = "domain"//"user"
	// permisn.Domain = "www.roleyzhang.com"

	// permision, err := utils.StartSrv(drive.DriveScope).
	// 	Permissions.Create("1HDsJ0dWyP7_PFTV7dNHlS_j5L_Fid65bL63j6O-Rn5U", permisn).Do()

	// if err != nil{
	// 	glog.Errorf("Unable to create share item: %v", err)
	// }

	// glog.V(8).Infof("ID: %s\n Role: %s\n Email: %s\n DisplayName: %s\n Domain: %s\n Type: %s\n",
	// 	permision.Id, permision.Role, permision.EmailAddress, permision.DisplayName, permision.Domain, permision.Type)

	// item, err := utils.StartSrv(drive.DriveScope).Permissions.
	// 	List("0B4_B23yaHaiYVkVlcGpOTl9WZVU").Do()

	// // 0B4_B23yaHaiYdk9uNWVYRlE3UkE
	// // 1wbmcPMimOknB5D4eQ9QuS6bXuFGlZ2B-

	// // 15b0-LD0DwN4Z7zs08ClVuHwCojcPUgKP
	// // 0B4_B23yaHaiYVkVlcGpOTl9WZVU
	// // Files.Get(id).
	// // // Files.Get("root").
	// // Fields("id, name, mimeType, parents, owners, createdTime").
	// // Do()
	// if err != nil {
	// 	println("shit happened: ", err.Error())
	// 	glog.Errorf("Unable to get share info: %v", err)
	// 	// return nil
	// }
	// for _, value := range item.Permissions {
	// 	glog.V(8).Infof("ID: %s\n allowDiscovery: %t\n displayName: %s\n domain: %s\n email: %s\n  expirationTime: %s\n  role: %s\n  type: %s\n",
	// 		value.Id, value.AllowFileDiscovery, value.DisplayName, value.Domain, value.EmailAddress, value.ExpirationTime, value.Role, value.Type)
	// 	for _, valu := range value.PermissionDetails {
	// 		glog.V(8).Infof("role: %s\n Inherited: %t\n PermissionType: %s\n InheritedFrom: %s\n",
	// 			valu.Role, valu.Inherited, valu.PermissionType, valu.InheritedFrom)
	// 	}
	// }
}

// func Lo() {
// 	//-----------------------------
// 	var total int64 = 1024 * 1024 * 1500
// 	reader := io.LimitReader(rand.Reader, total)

// 	p := mpb.New(
// 		mpb.WithWidth(60),
// 		mpb.WithRefreshRate(180*time.Millisecond),
// 	)

// 	bar := p.AddBar(total, mpb.BarStyle("[=>-|"),
// 		mpb.PrependDecorators(
// 			decor.CountersKibiByte("% .2f / % .2f"),
// 		),
// 		mpb.AppendDecorators(
// 			decor.EwmaETA(decor.ET_STYLE_GO, 90),
// 			decor.Name(" ] "),
// 			decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
// 		),
// 	)

// 	// create proxy reader
// 	proxyReader := bar.ProxyReader(reader)
// 	defer proxyReader.Close()

// 	// copy from proxyReader, ignoring errors
// 	_, err := io.Copy(ioutil.Discard, proxyReader)
// 	if err != nil{
// 		glog.V(8).Info(err)
// 	}

// 	p.Wait()
// 	//-----------------------------
// }
//-------------phase 6...
// var spinChars = `|/-\`

// type Spinner struct {
// 	message string
// 	i       int
// }

// func NewSpinner(message string) *Spinner {
// 	return &Spinner{message: message}
// }

// func (s *Spinner) Tick() {
// 	fmt.Printf("%s %c \r", s.message, spinChars[s.i])
// 	s.i = (s.i + 1) % len(spinChars)
// }

// func isTTY() bool {
// 	fi, err := os.Stdout.Stat()
// 	if err != nil {
// 		return false
// 	}
// 	return fi.Mode()&os.ModeCharDevice != 0
// }

// func Lo() {
// 	flag.Parse()
// 	s := NewSpinner("working...")
// 	isTTY := isTTY()
// 	// Ps.Pre = "                      [$1 $2]"
// 	for i := 0; i < 100; i++ {
// 		fmt.Printf("\rOn %d/10", i)
// 		if isTTY {
// 			s.Tick()
// 		}
// 		time.Sleep(100 * time.Millisecond)
// 	}
// 	Ps.SetPrefix(FixlongStringRunes(0),1)
// }

// func FixlongStringRunes(n int) string {
// 	b := make([]byte, n)
// 	for i := range b {
// 		b[i] = ' '
// 	}
// 	return string(b)
// }

//-------------phase 5...
//-------------phase 4
// func recursiveCall(id string, chF, chD chan string) {
// 	// glog.V(8).Info("B: ")
// 	// product += num

// 	// if num == 1 {
// 	//     ch <- product
// 	//     return
// 	// }
// 	// pthSep := string(os.PathSeparator)
// 	qString := "'" + id + "' in parents"
// 	glog.V(8).Info("B1: ", qString)
// 	item, err := utils.StartSrv(drive.DriveScope).Files.List().
// 		Q(qString).PageSize(40).
// 		Fields("nextPageToken, files(id, name, mimeType)").
// 		Do()
// 	// glog.V(8).Info("B2: ", item.Files)
// 	if err != nil {
// 		glog.Errorln("file or dir not exist: ", err.Error())
// 	}
// 	// glog.V(8).Info("B3: ")
// 	for _, file := range item.Files {
// 		// glog.V(8).Info("B4: ")
// 		if file.MimeType == "application/vnd.google-apps.folder" {
// 			// glog.V(8).Info("B5: ")
// 			chD <- file.Id + " : " + file.Name
// 			glog.V(8).Info("D----: ", file.Id, file.Name )
// 			go recursiveCall(file.Id, chF, chD)
// 		} else {
// 			// glog.V(8).Info("B6: ")
// 			chF <- file.Id + " : " + file.Name
// 			// glog.V(8).Info("F: ", file.Id, file.Name )
// 		}
// 	}
// 	// glog.V(8).Info("B7: ")

// }

// func Lo() {
// 	chF := make(chan string, 40)
// 	chD := make(chan string, 40)
// 	go recursiveCall("19YMYxawcjse0IcqKHrJYyx7yDEA_SLEA", chF, chD)
// 	for n := range chF {
// 		go func(n string) {
// 			// file := <-n
// 			glog.V(8).Info("F: ", n)
// 		}(n)

// 	}
// 	for n := range chD {
// 		go func(n string) {
// 			// file := <-n
// 			// glog.V(8).Info("F: ", n)
// 			glog.V(8).Info("D: ", n)
// 		}(n)
// 	}

// 	// close(chF)
// 	// close(chD)
// 	// for{
// 	// 	file := <-chF
// 	// 	folder := <-chD
// 	// 	glog.V(8).Info("F: ", file , " D: ", folder)
// 	// 	// close(chF)
// 	// 	// close(chD)
// 	// }
// }

//-------------phase 3
// func downloader(id int, c chan string) {
// 	for n := range c {
// 		time.Sleep(time.Second)
// 		if id ==0 {
// 			glog.V(8).Infoln("Receiced cd: ", n)
// 		}
// 		if id ==1 {
// 			glog.V(8).Infoln("Receiced cf: ", n)

// 		}
// 	}
// }

// func createDownloader(id int) chan<- string{
// 	c := make(chan string)
// 	go downloader(id, c)
// 	return c
// }

// func checkDrvData(id string) (c1, c2 chan string) {
// 	cd := make(chan string,40)
// 	cf := make(chan string,40)
// 	// c2 := make(chan string)
// 	qString := "'" + id + "' in parents"
// 	item, err := utils.StartSrv(drive.DriveScope).Files.List().
// 		Q(qString).PageSize(40).
// 		Fields("nextPageToken, files(id, name, mimeType)").
// 		Do()
// 	if err != nil {
// 		glog.Errorln("file or dir not exist: ", err.Error())
// 		// return nil
// 	}
// 	for _, file := range item.Files {
// 		if file.MimeType == "application/vnd.google-apps.folder" {
// 			// pat := path + pthSep + file.Name
// 			// glog.V(8).Info("D: ", pat)
// 			// folders = append(folders, pat)
// 			// if err != nil {
// 			// 	glog.Errorln("file or dir not exist: ", err.Error())
// 			// 	return nil, err
// 			// }
// 			glog.V(8).Info("D: ", file.Id+":"+file.Name)
// 			cd <- file.Id + ":" + file.Name
// 			go checkDrvData(file.Id )
// 		} else {
// 			// files = filesFromSrv(path, file.Id, file.Name)
// 			glog.V(8).Info("F: ", file.Id+":"+file.Name)
// 			Ps.SetPrefix(file.Id + ":" + file.Name)
// 			// files = append(files, path+pthSep+file.Name)
// 			cf <- file.Id + ":" + file.Name
// 		}
// 	}
// 	return cd, cf
// }

// func Lo(){

// 	// checkDrvData("19YMYxawcjse0IcqKHrJYyx7yDEA_SLEA")
// 	cd,cf := checkDrvData("19YMYxawcjse0IcqKHrJYyx7yDEA_SLEA")
// 	var downloader = createDownloader(0)
// 	var downloader2= createDownloader(1)

// 	var values []string
// 	var values2 []string
// 	var activeW chan <- string
// 	var activeV string
// 	var activeW2 chan <- string
// 	var activeV2 string
// 	if len(values) >0 {
// 		activeW = downloader
// 		activeV = values[0]
// 	}
// 	if len(values2) >0 {
// 		activeW2 = downloader2
// 		activeV2 = values2[0]
// 	}
// 	tick := time.Tick(time.Second)

// 	for {
// 		select {
// 		case n:= <-cd:
// 			glog.V(8).Infoln("Receiced cd: ", n)
// 			values = append(values, n)
// 		case n:= <-cf:
// 			glog.V(8).Infoln("Receiced cf: ", n)
// 			values2 = append(values2, n)
// 		case activeW <- activeV:
// 			values = values[1:]
// 		case activeW2 <- activeV2:
// 			values2 = values2[1:]
// 		case <-tick:
// 			fmt.Println(
// 				"queue len =", len(values), len(values2))

// 		}
// 	}
// }
// func Select() {
// 	Ps.SetPrefix("SELECT")
// // 	var c1, c2 chan int
// // 	// n1:= <- c1
// // 	// n2:= <- c2
// // 	select {
// // 	case n := <-c1:
// // 		glog.V(8).Info("receive from c1: ", n)
// // 	case n := <-c2:
// // 		glog.V(8).Info("receive from c2: ", n)
// // 	default:
// // 		glog.V(8).Info("receive from no one: ")

// // 	}

// }

/*
async download method
1. query function use go routine return task, if task failed then write into log
2. use select to handle query successful task, which query task return firstly, then run download task
*/
//-------------phase 2
// func doDownloader(id int, c chan int, wg *sync.WaitGroup) {
// 	for  n := range c {
// 		glog.V(8).Infof("Downloader %d receive %c\n", id, n)
// 		// go func(){done <- true}()
// 		Ps.SetPrefix(strconv.Itoa(n)+ strconv.Itoa(id))
// 		wg.Done()
// 	}
// }

// type downloader struct  {
// 	in chan int
// 	// done chan bool
// 	wg *sync.WaitGroup
// }

// // create downloader channels
// func createDownloader(id int, wg *sync.WaitGroup) downloader{
// 	w := downloader{
// 		in: make(chan int),
// 		wg: wg,
// 	}
// 	go doDownloader(id, w.in, wg)
// 	return w
// }

// // lo ...
// func Lo() {
// 	glog.V(8).Info("this is Lo Testing")
// 	var wg sync.WaitGroup
// 	wg.Add(20)
// 	var downloaders [10]downloader
// 	for i := 0; i < 10; i++ {
// 		// channels[i] = make(chan int)
// 		// go downloader(i, channels[i])
// 		downloaders[i] = createDownloader(i, &wg)
// 	}
// 	for i, downloader := range downloaders {
// 		downloader.in <- 'a' + i
// 	}
// 	for i, downloader := range downloaders {
// 		downloader.in <- 'A' + i
// 	}
// 	wg.Wait()
// }

//-------------phase 1
// func downloader(id int, c chan int) {
// 	// for {
// 	// 	// n := <-c
// 	// 	glog.V(8).Infof("Downloader %d receive %c\n", id, <-c)
// 	// }
// 	for  n := range c {
// 		// n, ok := <-c
// 		// if !ok {
// 		// 	break
// 		// }
// 		glog.V(8).Infof("Downloader %d receive %c\n", id, n)
// 	}
// }

// // create downloader channels
// func createDownloader(id int) chan int {
// 	c := make(chan int)
// 	// go func() {
// 	// 	for {
// 	// 		glog.V(8).Infof("Downloader %d receive %c\n", id, <-c)
// 	// 	}
// 	// }()
// 	go downloader(id, c)
// 	return c
// }

// // create bufferedChannel
// func BufferedChannel() {
// 	glog.V(8).Infof("BufferedChannel:")
// 	c := make(chan int, 3)
// 	go downloader(0, c)
// 	c <- 'a'
// 	c <- 'b'
// 	c <- 'c'
// 	c <- 'd'
// 	close(c)
// }

// // lo ...
// func Lo() {
// 	glog.V(8).Info("this is Lo Testing")
// 	var channels [10]chan int
// 	for i := 0; i < 10; i++ {
// 		// channels[i] = make(chan int)
// 		// go downloader(i, channels[i])
// 		channels[i] = createDownloader(i)
// 	}
// 	for i := 0; i < 10; i++ {
// 		channels[i] <- 'a' + i
// 	}
// 	for i := 0; i < 10; i++ {
// 		channels[i] <- 'A' + i
// 	}
// 	// c := make(chan int)
// 	// c <- 1
// 	// c <- 2
// 	time.Sleep(time.Millisecond)
// }
