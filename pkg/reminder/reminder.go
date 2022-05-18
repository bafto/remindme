// package reminder handles the interaction
// with saved data
package reminder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const (
	TimeLayout = time.UnixDate
)

var (
	savePath string
)

// make sure the save file exists
func init() {
	if path, err := os.UserCacheDir(); err == nil {
		savePath = filepath.Join(path, "/remindme/remindme.json")
	} else {
		log.Fatal(err)
	}

	if _, err := os.Stat(savePath); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(savePath, []byte("[]"), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
	f, err := os.OpenFile(savePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if content, err := io.ReadAll(f); err != nil {
		log.Fatal(err)
	} else if !json.Valid(content) {
		fmt.Printf("The cache file contains invalid json data.\nDo you want to clear the file? [y/n]: ")
		var s string
		fmt.Scanf("%s", &s)
		if s[0] != 'y' {
			os.Exit(1)
		}
		if err := os.WriteFile(savePath, []byte("[]"), os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
}

// represents a single entry in the saved reminders
type Entry struct {
	When  string
	Title string
	Msg   string
	Id    uuid.UUID
}

func NewEntry(when time.Time, title, msg string) Entry {
	return Entry{
		When:  when.Format(TimeLayout),
		Title: title,
		Msg:   msg,
		Id:    uuid.New(),
	}
}

func (e *Entry) GetTime() (time.Time, error) {
	return time.Parse(TimeLayout, e.When)
}

func overrideReminders(entries []Entry) error {
	data, err := json.MarshalIndent(entries, "", "\t")
	if err != nil {
		return err
	}

	f, err := os.OpenFile(savePath, os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

func GetAllReminders() (entries []Entry, err error) {
	f, err := os.OpenFile(savePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&entries)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func AddReminder(entry Entry) error {
	if entries, err := GetAllReminders(); err != nil {
		return err
	} else {
		entries = append(entries, entry)
		return overrideReminders(entries)
	}
}

// removes the specified entry by comparing it
// returns the number of remaining reminders
func RemoveReminder(entry Entry) (int, error) {
	entries, err := GetAllReminders()
	if err != nil {
		return 0, err
	}

	// remove the specified entry by checking
	for i := 0; i < len(entries); i++ {
		if entries[i].Id == entry.Id {
			entries[i] = entries[len(entries)-1]
			entries = entries[:len(entries)-1]
			i--
		}
	}

	return len(entries), overrideReminders(entries)
}
