package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
  	"os/signal"
  	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
  	"io"

	"github.com/guregu/null"
	"github.com/hugolgst/rich-go/client"
	"github.com/kardianos/osext"
	"github.com/pelletier/go-toml"
	"github.com/reujab/wallpaper"
)

type WalltakerData struct {
	ID               int         `json:"id"`
	Expires          time.Time   `json:"expires"`
	UserID           int         `json:"user_id"`
	Terms            string      `json:"terms"`
	Blacklist        string      `json:"blacklist"`
	PostURL          null.String `json:"post_url"`
	PostThumbnailURL interface{} `json:"post_thumbnail_url"`
	PostDescription  interface{} `json:"post_description"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
    SetBy            null.String `json:"set_by"`
	URL              string      `json:"url"`
}

func getWalltakerData(url string) WalltakerData {
	webClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "Walltaker Go Client/1.1.1")

	res, getErr := webClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	userData := WalltakerData{}
	jsonErr := json.Unmarshal(body, &userData)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return userData
}

type NoDataError struct {
	IntA int
	IntB int
	Msg  string
}

func (e *NoDataError) Error() string {
	return e.Msg
}

func getWallpaperUrlFromData(userData WalltakerData) (string, error) {
	if userData.PostURL.String == "" {
		return "", &NoDataError{
			Msg: fmt.Sprintf("No data found for ID %d", userData.ID),
		}
	}
	return userData.PostURL.String, nil
}

func clearWindowsWallpaperCache() {
	// Remove cached wallpaper files, issue #12
	if runtime.GOOS == "windows" {
		windowsWallpaperCacheDir := os.Getenv("APPDATA") + "\\Microsoft\\Windows\\Themes"
		if _, err := os.Stat(windowsWallpaperCacheDir + "\\TranscodedWallpaper"); !os.IsNotExist(err) {
			e := os.Remove(windowsWallpaperCacheDir + "\\TranscodedWallpaper")
			if e != nil {
				log.Fatal(e)
			}
		}
		if _, err2 := os.Stat(windowsWallpaperCacheDir + "\\CachedFiles"); !os.IsNotExist(err2) {
			e2 := os.RemoveAll(windowsWallpaperCacheDir + "\\CachedFiles")
			if e2 != nil {
				log.Fatal(e2)
			}
		}
	}
}

func goSetWallpaper(url string, saveLocally bool) {
	clearWindowsWallpaperCache()
 	if runtime.GOOS != "windows" {
		wallpaper.SetFromFile("") // free up for macOS
	}
 	err := wallpaper.SetFromURL(url)

	 if saveLocally {
  		saveWallpaperLocally(url)
 	}

	 if err != nil {
  		fmt.Println("Ouch! Had a problem while setting your wallpaper.")
    		fmt.Println("Full error: ", err)
  	}
	return
}

func saveWallpaperLocally(url string) {
	folderPath, err := osext.ExecutableFolder()
 	filename := filepath.Join(folderPath, "download", path.Base(url))
 	_, err = os.Stat(filename)

	 if os.IsNotExist(err) {

		//fmt.Printf("Downloading", url, " to ", filename)
  		response, err := http.Get(url)
  		if err != nil {
  			return
  		}

        defer response.Body.Close()

		file, err := os.Create(filename)
		if err != nil {
  			return
  		}
  		defer file.Close()
 		_, err = io.Copy(file, response.Body)
 	} else {
  		fmt.Printf("Wallpaper file already exists, skipping! ")
	}
	return
}

func main() {
	// fmt.Println("WALLTAKER CLIENT")
	fmt.Println(`
	██╗    ██╗ █████╗ ██╗     ██╗  ████████╗ █████╗ ██╗  ██╗███████╗██████╗
	██║    ██║██╔══██╗██║     ██║  ╚══██╔══╝██╔══██╗██║ ██╔╝██╔════╝██╔══██╗
	██║ █╗ ██║███████║██║     ██║     ██║   ███████║█████╔╝ █████╗  ██████╔╝
	██║███╗██║██╔══██║██║     ██║     ██║   ██╔══██║██╔═██╗ ██╔══╝  ██╔══██╗
	╚███╔███╔╝██║  ██║███████╗███████╗██║   ██║  ██║██║  ██╗███████╗██║  ██║
	 ╚══╝╚══╝ ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝

	 	v1.1.1. Go client by @OddPawsX
	 		 	Walltaker by Gray over at joi.how <3

	(You can minimize this window; it will periodically check in for new wallpapers)
	`)

	folderPath, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Loaded config from " + filepath.Join(folderPath, "walltaker.toml"))

	dat, err := os.ReadFile(filepath.Join(folderPath, "walltaker.toml"))
	if err != nil {
		log.Fatal(err)
	}

	bg, err := wallpaper.Get()
 	fmt.Println("Detected original wallpaper as: ", bg)

	c := make(chan os.Signal)
 	signal.Notify(c, os.Interrupt)
 	go func() {
  		<-c
		wallpaper.SetFromFile(bg)
		os.Exit(0)
 	}()

	tomlDat := string(dat)

	config, _ := toml.Load(tomlDat)

	base := config.Get("Base.base").(string)
	feed := config.Get("Feed.feed").(int64)
	freq := config.Get("Preferences.interval").(int64)
	mode := config.Get("Preferences.mode").(string)
	saveLocally := config.Get("Preferences.saveLocally").(bool)
	useDiscord := config.Get("Preferences.discordPresence").(bool)

	builtUrl := base + strconv.FormatInt(feed, 10) + ".json"

	if useDiscord == true {
		discorderr := client.Login("942796233033019504")
		if discorderr != nil {
			log.Fatal(discorderr)
		}

		timeNow := time.Now()
		discorderr = client.SetActivity(client.Activity{
			State:      "Set my wallpaper~",
			Details:    strings.Replace(builtUrl, ".json", "", -1),
			LargeImage: "eggplant",
			LargeText:  "Powered by joi.how",
			Timestamps: &client.Timestamps{
				Start: &timeNow,
			},
		})

		if discorderr != nil {
			log.Fatal(discorderr)
		}
  	}

	if saveLocally == true {
  		fmt.Println("Local saving enabled")
    		_, err := os.Stat(filepath.Join(folderPath, "download"))
		if os.IsNotExist(err) {
			fmt.Println("Created download directory since it did not exist")
			os.Mkdir(filepath.Join(folderPath, "download"), os.FileMode(0777))
    		}
  	}

	fmt.Printf("Checking in every %d seconds...\r\n", freq)

	userData := getWalltakerData(builtUrl)

	wallpaperUrl, noDataErr := getWallpaperUrlFromData(userData)
	ready := noDataErr == nil
	for ready == false {
		if noDataErr != nil {
			// log.Fatal(noDataErr)
			fmt.Printf("No data for ID %d, trying again in %d seconds...\r\n", feed, freq)
			time.Sleep(time.Second * time.Duration(freq))
			userData = getWalltakerData(builtUrl)
			wallpaperUrl, noDataErr = getWallpaperUrlFromData(userData)
		} else {
			ready = true
		}
	}

 	goSetWallpaper(wallpaperUrl, saveLocally)
 	fmt.Println("Set initial wallpaper: DONE")

	if strings.ToLower(mode) == "fit" {
		err = wallpaper.SetMode(wallpaper.Fit)
	} else if strings.ToLower(mode) == "crop" {
		err = wallpaper.SetMode(wallpaper.Crop)
	} else {
		err = wallpaper.SetMode(wallpaper.Crop)
	}

	oldWallpaperUrl := wallpaperUrl

	for range time.Tick(time.Second * time.Duration(freq)) {
		fmt.Printf("Polling... ")
		userData := getWalltakerData(builtUrl)
		wallpaperUrl := userData.PostURL.String
        setterName := userData.SetBy.String

		if wallpaperUrl != oldWallpaperUrl {
            if setterName != "" {
               fmt.Printf(setterName)
               fmt.Printf(" set your wallpaper! Setting... ")
            } else {
              fmt.Printf("New wallpaper found! Setting... ")
            }
      		goSetWallpaper(wallpaperUrl, saveLocally)
    		fmt.Printf("Set!")
			oldWallpaperUrl = wallpaperUrl
		} else {
			fmt.Printf("Nothing new yet.")
		}
		fmt.Printf("\r\n")
	}
}
