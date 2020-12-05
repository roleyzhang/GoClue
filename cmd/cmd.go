package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/cheynewallace/tabby"
	"github.com/dustin/go-humanize"
	"github.com/golang/glog"
	"github.com/logrusorgru/aurora"
	"github.com/roleyzhang/GoClue/utils"
	"github.com/theckman/yacspin"
	"go.uber.org/ratelimit"
	"golang.org/x/net/context"
	"google.golang.org/api/drive/v3"
)

var qString string

// DirSug ... for store remote folder prompt
var DirSug *[]prompt.Suggest

// Page ... store nextpagetoken, key is int, value string
var Page map[int]string

// FileSug ... for store remote file prompt
var FileSug *[]prompt.Suggest

// PathSug ... for store local path prompt
var PathSug *[]prompt.Suggest

// AllSug ... for store files & folders prompt
var AllSug *[]prompt.Suggest

// IdfileSug ... for store file id prompt
var IdfileSug *[]prompt.Suggest

// IddirSug ... for store folder id prompt
var IddirSug *[]prompt.Suggest

// IDAllSug ... for store files & folders id prompt
var IDAllSug *[]prompt.Suggest

// TypesSug ... for store remote file type prompt
var TypesSug *[]prompt.Suggest

// RoleSug ... for store role prompt
var RoleSug *[]prompt.Suggest

// GmailSug ... for store gmail prompt
var GmailSug *[]prompt.Suggest

// DomainSug ... for store domain prompt
var DomainSug *[]prompt.Suggest

// CommentSug ... for store commnet prompt
var CommentSug *[]prompt.Suggest

// CmtListSug ... for store comment prompt
var CmtListSug *[]prompt.Suggest

// Ps ... prompt style
var Ps PromptStyle

// Ii ... struct of item information
var Ii ItemInfo

var cfg *yacspin.Config
var pthSep string
var colorYellow string
var colorRed string
var commands map[string]string
var perPageSize int64

// ItemInfo ... struct
type ItemInfo struct {
	// item       *drive.File
	Path         map[string]string
	RootID       string
	ItemID       string
	DeleteItemIs string
	maxLength    int
}

// PromptStyle ... struct
type PromptStyle struct {
	Pre      string
	Gap      string
	FolderID string
	Info     string
	Status   string
}

func init() {
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

	comment := []prompt.Suggest{
		{Text: "-c", Description: "create comment"},
		{Text: "-d", Description: "delete comment"},
		{Text: "-u", Description: "update comment"},
		{Text: "-l", Description: "list comment"},
		{Text: "-g", Description: "get comment"},
	}
	CommentSug = &comment

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
		Path:         make(map[string]string),
		RootID:       "",
		ItemID:       "",
		DeleteItemIs: "",
		maxLength:    40,
	}

	Ps = PromptStyle{
		Pre:      "$0[$1 $2]",
		Gap:      ">>>",
		FolderID: "",
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

// getSugID ...
func (ii *ItemInfo) getSugID(sug *[]prompt.Suggest, text string) (string, error) {
	// fmt.Println(text)
	if sug != nil {
		for _, v := range *sug {
			if v.Text == text {
				return v.Description, nil
			}
		}
	}
	qString := "name='" + text + "'" + " and '" + ii.ItemID + "' in parents " + " and trashed=false"

	file, err := utils.StartSrv(drive.DriveScope).Files.List().IncludeItemsFromAllDrives(true).IncludeTeamDriveItems(true).SupportsAllDrives(true).
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

// SetPrefix ... set prompt prefix
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
}

// SetDynamicPrefix ... set promt dynamic info
func (ps *PromptStyle) SetDynamicPrefix() (string, bool) {
	// glog.V(8).Info("SetPrefix: ",msgs, ii.ItemId, len(*DirSug) )
	var result string
	if DirSug != nil {
		folderName := GetSugDec(DirSug, ps.FolderID)
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

// GetRoot ... get remote home folder
func (ps *PromptStyle) GetRoot(ii *ItemInfo) {
	dirInfo := utils.GetSugInfo()
	item, err := utils.StartSrv(drive.DriveScope).
		// Files.Get(id).
		Files.Get("root").SupportsAllDrives(true).SupportsTeamDrives(true).
		Fields("id, name, mimeType").
		Do()
	if err != nil {
		glog.Errorf("Unable to retrieve root: %v", err)
	}
	if item.MimeType == "application/vnd.google-apps.folder" {
		ii.Path[item.Id] = item.Name
		ii.ItemID = item.Id
		ii.RootID = item.Id
		ps.FolderID = item.Id
		// setting the prompt.Suggest
		s2 := prompt.Suggest{Text: item.Name, Description: item.Id}
		DirSug = dirInfo(s2, ii.DeleteItemIs, 0)
	}
}

//-------------------------------
// breakDown ...
func breakDown(path string) []string {
	return strings.Split(path, "/")
}

// ShowResult ... print the request result
func (ii *ItemInfo) ShowResult(
	page map[int]string,
	counter int,
	param, cmd, scope string) *drive.FileList { // This should testing by change the authorize token
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
		DirSug = dirInfo(s, ii.DeleteItemIs, 0)

	}
	if param != "next" && param != "previous" {
		qString = commands[param]
		if strings.Contains(qString, "$") {
			qString = strings.ReplaceAll(qString, "$", cmd)
		}
	}
	if param == "dir" {
		iD, err := ii.getSugID(AllSug, cmd)
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
	r, err := utils.StartSrv(scope).Files.List().IncludeItemsFromAllDrives(true).IncludeTeamDriveItems(true).SupportsAllDrives(true).
		// r, err := utils.StartSrv(scope).Files.List().IncludeItemsFromAllDrives(true).SupportsAllDrives(true).
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
				var name string
				var types string
				if len(i.Name) > ii.maxLength {
					name = i.Name[:ii.maxLength] + "..."
				} else {
					name = i.Name
				}
				if len(i.Permissions) > 1 {
					types = fmt.Sprintf("%s %s", aurora.Brown("*S*"), strings.Split(i.MimeType, "/")[1])
				} else {
					types = strings.Split(i.MimeType, "/")[1]
				}
				var onm string
				if len(i.Owners) > 0 {
					onm = i.Owners[0].DisplayName
				} else {
					onm = i.DriveId
				}
				t.AddLine(aurora.Bold(aurora.Cyan(name)),
					aurora.Bold(aurora.Cyan(i.Id)),
					aurora.Bold(aurora.Cyan(types)),
					// aurora.Bold(aurora.Cyan(i.Owners[0].DisplayName)),
					aurora.Bold(aurora.Cyan(onm)),
					aurora.Bold(aurora.Cyan(i.CreatedTime)))
				s := prompt.Suggest{Text: i.Id, Description: i.Name}
				s2 := prompt.Suggest{Text: i.Name, Description: i.Id}
				DirSug = dirInfo(s2, ii.DeleteItemIs, 0)
				AllSug = allInfo(s2, ii.DeleteItemIs, 0)
				IddirSug = iddirInfo(s, ii.DeleteItemIs, 1)
				IDAllSug = idAllInfo(s, ii.DeleteItemIs, 1)
			} else {
				var name string
				var types string
				if len(i.Name) > ii.maxLength {
					name = i.Name[:ii.maxLength] + "..."
				} else {
					name = i.Name
				}
				if len(i.Permissions) > 1 {
					types = fmt.Sprintf("%s %s", aurora.Brown("*S*"), strings.Split(i.MimeType, "/")[1])
				} else {
					types = strings.Split(i.MimeType, "/")[1]
				}
				if i.MimeType == "application/vnd.google-apps.shortcut" {
					t.AddLine(aurora.Cyan(name),
						aurora.Cyan(i.Id),
						aurora.Cyan(types),
						aurora.Cyan(i.Owners[0].DisplayName),
						aurora.Cyan(i.CreatedTime))

				} else {
					var onm string
					if len(i.Owners) > 0 {
						onm = i.Owners[0].DisplayName
					} else {
						onm = i.DriveId
					}
					t.AddLine(aurora.Green(name),
						aurora.Green(i.Id),
						aurora.Green(types),
						aurora.Green(onm),
						aurora.Green(i.CreatedTime))
				}
				s := prompt.Suggest{Text: i.Id, Description: i.Name}
				s2 := prompt.Suggest{Text: i.Name, Description: i.Id}
				FileSug = fileInfo(s2, ii.DeleteItemIs, 0)
				AllSug = allInfo(s2, ii.DeleteItemIs, 0)
				IdfileSug = idfileInfo(s, ii.DeleteItemIs, 1)
				IDAllSug = idAllInfo(s, ii.DeleteItemIs, 1)
			}
		}
		t.Print()
	}
	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
	// fmt.Println("pathlen", len(*PathSug))
	return r
}

func (ii *ItemInfo) setSult(id, scope string) { // This should testing by change the authorize token

	dirInfo := utils.GetSugInfo()
	fileInfo := utils.GetSugInfo()
	allInfo := utils.GetSugInfo()
	idfileInfo := utils.GetSugInfo()
	iddirInfo := utils.GetSugInfo()
	idAllInfo := utils.GetSugInfo()

	//--------every time runCommand add folder history to dirSug
	for key, value := range ii.Path {
		s := prompt.Suggest{Text: value, Description: key}
		DirSug = dirInfo(s, ii.DeleteItemIs, 0)

	}
	qString := "'" + id + "' in parents and trashed=false"
	glog.V(5).Infoln("qString: ", qString)
	r, err := utils.StartSrv(scope).Files.List().IncludeItemsFromAllDrives(true).IncludeTeamDriveItems(true).SupportsAllDrives(true).
		Q(qString).
		PageSize(40).
		Fields("nextPageToken, files(id, name, mimeType )").
		Do()

	if err != nil {
		// uncomment below will cause 500 error and program exit why?
		glog.Errorf("Unable to retrieve files: %v", err)
	}
	if len(r.Files) == 0 {
		glog.V(1).Info("No files found.")
	} else {
		for _, i := range r.Files {
			if i.MimeType == "application/vnd.google-apps.folder" {
				s := prompt.Suggest{Text: i.Id, Description: i.Name}
				s2 := prompt.Suggest{Text: i.Name, Description: i.Id}
				DirSug = dirInfo(s2, ii.DeleteItemIs, 0)
				AllSug = allInfo(s2, ii.DeleteItemIs, 0)
				IddirSug = iddirInfo(s, ii.DeleteItemIs, 1)
				IDAllSug = idAllInfo(s, ii.DeleteItemIs, 1)
			} else {
				s := prompt.Suggest{Text: i.Id, Description: i.Name}
				s2 := prompt.Suggest{Text: i.Name, Description: i.Id}
				FileSug = fileInfo(s2, ii.DeleteItemIs, 0)
				AllSug = allInfo(s2, ii.DeleteItemIs, 0)
				IdfileSug = idfileInfo(s, ii.DeleteItemIs, 1)
				IDAllSug = idAllInfo(s, ii.DeleteItemIs, 1)
			}
		}
	}
}

// PathGenerate ... generate folder path...
func PathGenerate(path, level string) {

	pathInfo := utils.GetLocalPathInfo()
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

// PathFileGenerate ... generate folder and file path...
func PathFileGenerate(path, level string) {
	pathInfo := utils.GetLocalPathInfo()
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
	}
}

// SetPrefix ...

// Rmd ... delete file by id
func (ii *ItemInfo) Rmd(id string) error {
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
	msgs := fmt.Sprintf("   Delete %s... %s ", id, aurora.Brown("done"))
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
	_, err = utils.StartSrv(drive.DriveScope).Files.Get(id).Do()
	if err != nil {
		spinner.StopFailMessage("   file or dir not exist")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("file or dir not exist", err)
		}
		glog.Error("file or dir not exist: ", err.Error())
		return err
	}

	if id == ii.RootID {
		spinner.StopFailMessage("   The root folder should not be deleted")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("The root folder should not be deleted", err)
		}
		return errors.New("The root folder should not be deleted")
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
	ii.DeleteItemIs = strings.TrimSuffix(id, " ")
	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
	return nil
}

// Rm ... delete file
func (ii *ItemInfo) Rm(name string) error {
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
	msgs := fmt.Sprintf("   Delete %s... %s ", name, aurora.Brown("done"))
	spinner.StopMessage(msgs)
	err = spinner.CharSet(yacspin.CharSets[9])
	// handle the error
	if err != nil {
		glog.V(8).Info("Spin run error", err.Error())
	}

	if err := spinner.Start(); err != nil {
		glog.V(8).Info("Spin start error", err.Error())
	}
	//-----yacspin-----------------
	//TODO: delete file
	var id string
	iD, err := ii.getSugID(DirSug, strings.TrimSuffix(name, " "))
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

	_, err = utils.StartSrv(drive.DriveScope).Files.Get(id).Do()
	if err != nil {
		glog.Errorln("file or dir not exist: ", err.Error())
		spinner.StopFailMessage("   file or dir not exist")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("file or dir not exist", err)
		}
		return err
	}

	if id == ii.RootID {
		spinner.StopFailMessage("   The root folder should not be deleted")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("The root folder should not be deleted", err)
		}
		return errors.New("The root folder should not be deleted")
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
	ii.DeleteItemIs = id
	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
	return nil
}

// Trash ... put file or folder to remote trash
func (ii *ItemInfo) Trash(name string) error {
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
	err = spinner.CharSet(yacspin.CharSets[9])
	// handle the error
	if err != nil {
		glog.V(8).Info("Spin run error", err.Error())
	}

	if err := spinner.Start(); err != nil {
		glog.V(8).Info("Spin start error", err.Error())
	}
	//-----yacspin-----------------
	//TODO: trash file
	var id string
	iD, err := ii.getSugID(DirSug, strings.TrimSuffix(name, " "))
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

	if id == ii.RootID {
		return errors.New("The root folder should not be trashed")
	}

	_, err = utils.StartSrv(drive.DriveScope).Files.Update(file.Id, &drive.File{Trashed: true}).Do()

	if err != nil {
		glog.Errorln("file or dir trashed failed: ", err.Error())
		return err
	}
	// fmt.Printf(string(colorYellow), file.Name, "", "Be Trashed")
	msgs := fmt.Sprintf("   Trash %s... %s ", file.Name, aurora.Brown("done"))
	spinner.StopMessage(msgs)
	ii.DeleteItemIs = id
	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
	return nil
}

// Trashd ... trash file or folder by id...
func (ii *ItemInfo) Trashd(id string) error {
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
	err = spinner.CharSet(yacspin.CharSets[9])
	// handle the error
	if err != nil {
		glog.V(8).Info("Spin run error", err.Error())
	}

	if err := spinner.Start(); err != nil {
		glog.V(8).Info("Spin start error", err.Error())
	}
	//-----yacspin-----------------
	file, err := utils.StartSrv(drive.DriveScope).Files.Get(id).Do()
	if err != nil {
		glog.Errorln("file or dir not exist: ", err.Error())
		return err
	}

	if id == ii.RootID {
		return errors.New("The root folder should not be trashed")
	}

	file, err = utils.StartSrv(drive.DriveScope).Files.Update(file.Id, &drive.File{Trashed: true}).Do()

	if err != nil {
		glog.Errorln("file or dir trashed failed: ", err.Error())
		return err
	}
	// fmt.Printf(string(colorYellow), file.Name, "", "Be Trashed")
	msgs := fmt.Sprintf("   Trash %s... %s ", file.Name, aurora.Brown("done"))
	ii.DeleteItemIs = strings.TrimSuffix(id, " ")
	spinner.StopMessage(msgs)
	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
	return nil
}

func (ii *ItemInfo) upload(file, parentID string) (*drive.File, error) {
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
	msgs := fmt.Sprintf("... %s ", aurora.Brown("done"))
	spinner.StopMessage(msgs)
	err = spinner.CharSet(yacspin.CharSets[71])
	// handle the error
	if err != nil {
		glog.V(8).Info("Spin run error", err.Error())
	}

	if err := spinner.Start(); err != nil {
		glog.V(8).Info("Spin start error", err.Error())
	}
	//-----yacspin-----------------
	// ii.DeleteItemIs = ""
	if parentID == "" {
		parentID = ii.ItemID
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
		Parents: []string{parentID},
	}
	// ufile, err := startSrv(drive.DriveScope).Files.Create(u).Media(fi).Do()
	ufile, err := utils.StartSrv(drive.DriveScope).Files.
		Create(u).
		ResumableMedia(context.Background(), fi, fileInfo.Size(), "").
		ProgressUpdater(func(now, size int64) {
			// fmt.Printf("%d, %d\r", now, size)
			rate := now * 100 / size
			mesg := fmt.Sprintf("%d%%  %s / %s",
				aurora.Brown(rate),
				aurora.Brown(humanize.Bytes(uint64(now))),
				aurora.Brown(humanize.Bytes(uint64(size))))
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

// CreateDir ... create folder
func (ii *ItemInfo) CreateDir(name string) (*drive.File, error) {
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
	err = spinner.CharSet(yacspin.CharSets[9])
	// handle the error
	if err != nil {
		glog.V(8).Info("Spin run error", err.Error())
	}

	if err := spinner.Start(); err != nil {
		glog.V(8).Info("Spin start error", err.Error())
	}
	//-----yacspin-----------------
	// ii.DeleteItemIs = ""
	d := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{ii.ItemID},
	}
	dir, err := utils.StartSrv(drive.DriveScope).Files.Create(d).Do()

	if err != nil {
		glog.Errorln("Could not create dir: ", err.Error())
		return nil, err
	}

	// fmt.Printf(string(colorYellow), dir.Name, "", " has been created")
	msgs := fmt.Sprintf("   Create %s... %s ", dir.Name, aurora.Brown("done"))
	spinner.StopMessage(msgs)
	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
	return dir, nil
}

// CreateInDir ...
func CreateInDir() func(name, path, parentId string) (map[string]string, *drive.File, error) {
	// //-----yacspin-----------------
	// spinner, err := yacspin.New(*cfg)
	// if err != nil {
	// 	glog.Error("Spin run error", err.Error())
	// }
	// if err := spinner.Frequency(100 * time.Millisecond); err != nil {
	// 	glog.Error("Spin run error", err.Error())
	// }
	// // msg := fmt.Sprintf("   Getting %s information from server", id)
	// // spinner.Suffix(msg)
	// err = spinner.CharSet(yacspin.CharSets[9])
	// // handle the error
	// if err != nil {
	// 	glog.V(8).Info("Spin run error", err.Error())
	// }

	// if err := spinner.Start(); err != nil {
	// 	glog.V(8).Info("Spin start error", err.Error())
	// 	// glog.Errorf("Spin start error %v", err)
	// 	// return err
	// }
	// //-----yacspin-----------------
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

		// msgs := fmt.Sprintf("   Create %s... %s ", dir.Name, Brown("done"))
		// spinner.StopMessage(msgs)
		parents[dir.Name] = dir.Id
		return parents, dir, nil
	}
}

// GetNoded ... change folder by id
func (ii *ItemInfo) GetNoded(id string) {
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
	// println(id)
	item, err := utils.StartSrv(drive.DriveScope).
		Files.Get(id).
		// Files.Get("root").
		Fields("id, name, mimeType, parents, owners, createdTime").
		Do()
	if err != nil {
		// println("shit happened: ", err.Error())
		spinner.StopFailMessage("   Unable to retrieve root")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("Unable to retrieve root", err)
		}
		glog.Errorf("Unable to retrieve root: %v", err)
		// return nil
	}
	if item != nil {
		if item.MimeType == "application/vnd.google-apps.folder" ||
			item.MimeType == "application/vnd.google-apps.shortcut" {
			ii.Path[item.Id] = item.Name
			ii.ItemID = item.Id
			if id == "root" {
				ii.RootID = item.Id
			}
			Ps.FolderID = item.Id
		}
	}
	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
}

// GetNode ... change folder
func (ii *ItemInfo) GetNode(cmd string) {
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
		iD, err := ii.getSugID(DirSug, strings.TrimSuffix(name, " "))
		if err != nil {
			spinner.StopFailMessage("   file or dir not exist")
			if err := spinner.StopFail(); err != nil {
				glog.V(8).Info("file or dir not exist", err)
			}
			glog.Errorln("file or dir not exist: " + err.Error())
			return
		}
		id = iD
	}
	item, err := utils.StartSrv(drive.DriveScope).
		Files.Get(id).SupportsAllDrives(true).SupportsTeamDrives(true).
		// Files.Get("root").
		Fields("id, name, mimeType, parents, owners, createdTime").
		Do()
	if err != nil {
		// println("shit happened: ", err.Error())
		spinner.StopFailMessage("   Unable to retrieve root")
		if err := spinner.StopFail(); err != nil {
			glog.V(8).Info("Unable to retrieve root", err)
		}
		glog.Errorf("Unable to retrieve root: %v", err)
		// return nil
	}
	if item != nil {
		if item.MimeType == "application/vnd.google-apps.folder" ||
			item.MimeType == "application/vnd.google-apps.shortcut" {
			ii.Path[item.Id] = item.Name
			ii.ItemID = item.Id
			if id == "root" {
				ii.RootID = item.Id
			}
			Ps.FolderID = item.Id

			// setting the prompt.Suggest
			s2 := prompt.Suggest{Text: item.Name, Description: item.Id}
			DirSug = dirInfo(s2, ii.DeleteItemIs, 0)
		}
	}
	ii.setSult(id, drive.DriveScope)
	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
}

// Move ... move file
func (ii *ItemInfo) Move(cmd string) error {
	//TODO: move file
	// println("this is .. move", cmd)
	if !strings.Contains(cmd, ">") {
		fmt.Printf(string(colorRed), "Wrong command format, please use \"h\" get help")
		return errors.New("Wrong command format, please use \"h\" get help")
	}
	fil := strings.Split(strings.Split(cmd, "mv ")[1], ">")
	iD, err := ii.getSugID(AllSug, strings.TrimSuffix(fil[0], " "))
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

	if file.Id == ii.RootID {
		return errors.New("The root folder should not be moved")
	}

	if len(breakDown(fil[1])) > 1 { // move to another folder
		newParentName := strings.Trim(breakDown(fil[1])[len(breakDown(fil[1]))-2], " ") // move to another folder
		newName := strings.Trim(breakDown(fil[1])[len(breakDown(fil[1]))-1], " ")       // change item name
		iD, err := ii.getSugID(AllSug, newParentName)
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

// PrintProgress ...
func (wc WriteCounter) PrintProgress() {
	rate := wc.Total * 100 / wc.Amount
	mesg := fmt.Sprintf("downloading  %d%%  %s / %s",
		aurora.Brown(rate), aurora.Brown(humanize.Bytes(wc.Total)), aurora.Brown(humanize.Bytes(wc.Amount)))
	wc.Spinner.Message(mesg)
	// msg("Downloading... "+ humanize.Bytes(wc.Total)+ " complete ")
}

// GetSugDec ... get suggest description
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
	fi := strings.Split(target, string(os.PathSeparator))
	msg := fmt.Sprintf("Downloading %s", fi[len(fi)-1])
	spinner.Suffix(msg)
	msgs := fmt.Sprintf("...to %s %s", path, aurora.Brown("done"))
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

// Download ... download file by name
func (ii *ItemInfo) Download(path, cmd string) error {
	//TODO: download file
	var id string
	if path == "" || cmd == "" {
		fmt.Printf(string(colorRed), "Wrong path format, please use \"h\" get help")
		return errors.New("Wrong path format, please use \"h\" get help")
	}
	// iD, err := ii.getSugId(AllSug, strings.TrimSuffix(fil[0], " "))
	iD, err := ii.getSugID(AllSug, strings.TrimSuffix(cmd, " "))
	if err != nil {
		fmt.Printf(string(colorRed), "file or dir not exist: "+err.Error())
		glog.Errorln("file or dir not exist: " + err.Error())
		return err
	}
	id = iD
	//2 start download
	targetPath := strings.Trim(path, " ") + string(os.PathSeparator) + strings.Trim(cmd, " ")
	StartingDownload(id, targetPath)
	return nil
}

// Downloadd file by id
func Downloadd(path, cmd string) error {
	//TODO: download file by id
	if path == "" || cmd == "" {
		fmt.Printf(string(colorRed), "Wrong path format, please input \"h\" get help")
		return errors.New("Wrong path format, please input \"h\" get help")
	}
	item, err := utils.StartSrv(drive.DriveScope).
		// Files.Get(id).
		Files.Get(cmd).SupportsAllDrives(true).SupportsTeamDrives(true).
		Fields("id, name, mimeType").
		Do()
	if err != nil {
		return err
	}
	if item.MimeType == "application/vnd.google-apps.folder" {
		glog.V(6).Info("B7: ", path)
		StartingDownload(cmd, path+string(os.PathSeparator)+item.Name)
	} else {
		glog.V(6).Info("B7: ", path)
		StartingDownload(cmd, path)
	}
	return nil
}

// GetAllDriveItems ... get all items from folder
func GetAllDriveItems(id, pageToken string, files *[]*drive.File) {

	// rl := ratelimit.New(5) // per second 5 requests
	// prev := time.Now()
	qString := "'" + id + "' in parents and trashed=false"
	// now := rl.Take()
	r, err := utils.StartSrv(drive.DriveScope).Files.List().IncludeItemsFromAllDrives(true).IncludeTeamDriveItems(true).SupportsAllDrives(true).
		Q(qString).
		// PageSize(perPageSize).
		PageSize(1000).
		Fields("nextPageToken, files(id, name, mimeType)").
		PageToken(pageToken).
		Do()
	if err != nil {
		glog.Errorln("file or dir not exist: ", err.Error())
	}
	*files = append(*files, r.Files...)
	glog.V(8).Info("files1 len:  ", len(*files))
	// if int64(len(r.Files)) == perPageSize {
	if int64(len(r.Files)) == 1000 {
		// time.Sleep(time.Second)
		pageToken := r.NextPageToken
		GetAllDriveItems(id, pageToken, files)
	}

	// now.Sub(prev)
	// prev = now
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
		msgs := fmt.Sprintf("... %s ", aurora.Brown("done"))
		spinner.StopMessage(msgs)
		err = spinner.CharSet(yacspin.CharSets[9])
		// handle the error
		if err != nil {
			glog.V(8).Info("Spin run error", err.Error())
		}

		if err := spinner.Start(); err != nil {
			glog.V(8).Info("Spin start error", err.Error())
		}
		//-----yacspin-----------------
		files := make([]*drive.File, 0)
		GetAllDriveItems(id, "", &files)
		if len(files) == 0 {
			glog.V(8).Info("--B2: ", len(files))
			fil, err := utils.StartSrv(drive.DriveScope).Files.Get(id).Do()
			if err != nil {
				glog.Error("file or dir not exist: ", err.Error())
			}
			switch fil.MimeType {
			case "application/vnd.google-apps.shortcut":
				spinner.StopFailMessage("   Google Drive Shortcut be downloaded")
				if err := spinner.StopFail(); err != nil {
					spinner.Stop()
					glog.Error("Shortcut cannot be downloaded: ", err)
					// stop <- "stop"
					// return
				}
				// stop <- "stop"
			case "application/vnd.google-apps.folder":
				spinner.StopFailMessage("   Empty folder be downloaded")
				if err := spinner.StopFail(); err != nil {
					spinner.Stop()
					glog.V(8).Info("Try to downloading empty folder", err)
					// stop <- "stop"
					// return
				}
				// stop <- "stop"
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
		} else {

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
					// glog.V(8).Info("B6-1: ", patt)
					pat = strings.Replace(pat, patt, "", 1)
					// glog.V(8).Info("B6-2: ", pat)
					out <- pat
				}
				now.Sub(prev)
				prev = now
			}
			if err := spinner.Stop(); err != nil {
				spinner.Stop()
				glog.Errorf("Spinner err: %v", err)
			}

		}
	}()
}

func downloader(id int, c chan string) {
	for n := range c {
		time.Sleep(time.Second)
		if strings.Contains(n, "-/-") {
			target := strings.Split(n, "-/-")
			glog.V(8).Infoln("------n >: ", target[3])
			// glog.V(8).Infoln("Starting downloading...", target[0], " ", target[1])
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
		}
	}
}

func createDownloader(id int) chan<- string {
	c := make(chan string)
	go downloader(id, c)
	return c
}

// StartingDownload ... starting download
func StartingDownload(id, path string) {
	glog.V(6).Info("download files: ", id, " : ", path)
	out := make(chan string)
	stop := make(chan string)

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
		case <-tick.C:
			// if buffer =0 and task count = full buffer size then jump out
			if len(values) == 0 && i == j {
				return
			}
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
			if !info.IsDir() && isDir {
				return nil
			}
			*files = append(*files, path)

		}
		return nil
	}
}

// GetLocalItems ... path is local path, noDir is switch list or folder
func GetLocalItems(path string, isDir bool, files *[]string) {
	// var files []string

	// fil := strings.Split(path, "u ")
	// err := filepath.Walk(fil[1], visit(files, isDir))
	err := filepath.Walk(path, visit(files, isDir))
	if err != nil {
		// panic(err)
		glog.Error(err)
	}
	// return files
}

// UpLod ... Upload function
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
	msgs := fmt.Sprintf("... %s ", aurora.Brown("done"))
	spinner.StopMessage(msgs)
	err = spinner.CharSet(yacspin.CharSets[29])
	// handle the error
	if err != nil {
		glog.V(8).Info("Spin run error", err.Error())
	}

	if err := spinner.Start(); err != nil {
		glog.V(8).Info("Spin start error", err.Error())
	}
	//-----yacspin-----------------
	// var tmpId string
	var parns map[string]string
	createInDir := CreateInDir()

	files := make([]string, 0)
	GetLocalItems(file, false, &files)
	for _, file := range files {
		// home, _ := os.UserHomeDir()
		dir, fil := filepath.Split(file)

		if utils.IsDir(file) {
			// glog.V(8).Info(ii.ItemId)
			qString := "name ='" + fil +
				"' and mimeType = 'application/vnd.google-apps.folder' and '" +
				ii.ItemID + "' in parents and trashed = false"
			// glog.V(5).Infoln("qString: ", qString)
			r, err := utils.StartSrv(scope).Files.List().IncludeItemsFromAllDrives(true).IncludeTeamDriveItems(true).SupportsAllDrives(true).
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
			}
			parents, _, err := createInDir(fil, dir, ii.ItemID)
			if err != nil {
				glog.Error(err)
				return
			}
			// tmpId = dr.Id
			parns = parents
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

// Share ... share files or folder
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
	msgs := fmt.Sprintf("...Share %s %s ", idorName, aurora.Brown("done"))
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
		iD, err := ii.getSugID(DirSug, strings.TrimSuffix(idorName, " "))
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

// Commnet ... comment for file or folder
func (ii *ItemInfo) Commnet(idorName, subcommand, content, updContent string, isByName bool) {
	cmtInfo := utils.GetSugInfo()
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
	// msgs := fmt.Sprintf("...Comment %s %s ", idorName, Brown("done"))
	// spinner.StopMessage(msgs)
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
		iD, err := ii.getSugID(DirSug, strings.TrimSuffix(idorName, " "))
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
	var comment *drive.Comment = new(drive.Comment)
	comment.Content = content
	switch subcommand {
	case "-c", "--c":
		cmt, errs := utils.StartSrv(drive.DriveScope).Comments.
			Create(id, comment).Fields("*").Do()

		if errs != nil {
			spinner.StopFailMessage("   Unable to comment item")
			if err := spinner.StopFail(); err != nil {
				glog.V(8).Info(" Unable to comment item", err)
			}
			glog.Errorf("Unable to comment item: %v", errs)
		}
		// glog.V(8).Info("commentid: ",cmt.Id, " : ", cmt.Content)
		msgs := fmt.Sprintf("...Create Comment %s %s ", cmt.Content, aurora.Brown("done"))
		spinner.StopMessage(msgs)
	case "-d", "--d":
		errs := utils.StartSrv(drive.DriveScope).Comments.
			Delete(id, content).Fields("*").Do()

		if errs != nil {
			spinner.StopFailMessage("   Unable to delete comment item")
			if err := spinner.StopFail(); err != nil {
				glog.V(8).Info(" Unable to delete comment item", err)
			}
			glog.Errorf("Unable to delete comment item: %v", errs)
		}
		msgs := fmt.Sprintf("...Delete Comment %s %s ", content, aurora.Brown("done"))
		spinner.StopMessage(msgs)
	case "-u", "--u":
		var comment *drive.Comment = new(drive.Comment)
		comment.Content = updContent
		_, errs := utils.StartSrv(drive.DriveScope).Comments.
			Update(id, content, comment).Fields("*").Do()

		if errs != nil {
			spinner.StopFailMessage("   Unable to comment item")
			if err := spinner.StopFail(); err != nil {
				glog.V(8).Info(" Unable to comment item", err)
			}
			glog.Errorf("Unable to comment item: %v", errs)
		}
		msgs := fmt.Sprintf("...Comment %s %s ", content, aurora.Brown("update done"))
		spinner.StopMessage(msgs)
	case "-l", "--l":
		cmt, errs := utils.StartSrv(drive.DriveScope).Comments.List(id).PageSize(100).Fields("*").Do()
		for _, value := range cmt.Comments {
			v := fmt.Sprintf("Comment ID: %s  Content: %s\n", aurora.Brown(value.Id), aurora.Brown(value.Content))
			fmt.Println(v)
			s := prompt.Suggest{Text: value.Id, Description: value.Content}
			CmtListSug = cmtInfo(s, ii.DeleteItemIs, 0)
		}
		if errs != nil {
			spinner.StopFailMessage("   Unable to list item comment")
			if err := spinner.StopFail(); err != nil {
				glog.V(8).Info(" Unable to list item comment", err)
			}
			glog.Errorf("Unable to list item comment: %v", errs)
		}
		msgs := fmt.Sprintf("...Comment %s %s ", id, aurora.Brown("list done"))
		spinner.StopMessage(msgs)
	case "-g", "--g":
		cmt, errs := utils.StartSrv(drive.DriveScope).Comments.Get(id, content).Fields("*").Do()

		v := fmt.Sprintf("Comment ID: %s  Content: %s  Author: %s  CreatedTime: %s\n",
			aurora.Brown(cmt.Id),
			aurora.Brown(cmt.Content),
			aurora.Brown(cmt.Author.DisplayName),
			aurora.Brown(cmt.CreatedTime))
		fmt.Println(v)
		for key, value := range cmt.Replies {
			v := fmt.Sprintf("Comment ID: %s  Content: %s\n",
				aurora.Brown(key),
				aurora.Brown(value))
			fmt.Println(v)

		}
		if errs != nil {
			spinner.StopFailMessage("   Unable to get item comment")
			if err := spinner.StopFail(); err != nil {
				glog.V(8).Info(" Unable to get item comment", err)
			}
			glog.Errorf("Unable to get item comment: %v", errs)
		}
		msgs := fmt.Sprintf("...Comment %s %s ", "detail ", aurora.Brown("list done"))
		spinner.StopMessage(msgs)
	}

	if err := spinner.Stop(); err != nil {
		glog.Errorf("Spinner err: %v", err)
	}
}

// Lo ... is a testing function
func Lo() {
}
