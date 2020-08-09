package cmd


import (
	"fmt"
	"github.com/c-bata/go-prompt"
)


var dirSug *[]prompt.Suggest
var fileSug *[]prompt.Suggest
var pathSug *[]prompt.Suggest
var allSug *[]prompt.Suggest
var idfileSug *[]prompt.Suggest
var iddirSug *[]prompt.Suggest

var LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}


type itemInfo struct {
	// item       *drive.File
	path   map[string]string
	rootId string
	itemId string
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


// SetPrefix ...
func (ii *itemInfo) SetPrefix(msgs string) {
	// folderId := ii.path[len(ii.path)-1]
	fmt.Println(ii.itemId)
	folderId := ii.itemId
	if dirSug != nil {
		folderName := getSugDec(dirSug, folderId)
		msg(folderName + msgs)
	}
}
