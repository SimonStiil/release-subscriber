package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/magiconair/properties"
)

var (
	arch = "amd64"
)

type TagList struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []*Tag `json:"results"`
}

type Tag struct {
	Creator         int64     `json:"creator"`
	ID              int64     `json:"id"`
	ImageID         string    `json:"image_id"`
	Images          []*Image  `json:"images"`
	LastUpdated     time.Time `json:"last_updated"`
	LastUpdater     int64     `json:"last_updater"`
	LastUpdaterUser string    `json:"last_updater_username"`
	Name            string    `json:"name"`
	Repository      int64     `json:"repository"`
	FullSize        int       `json:"full_size"`
	V2              bool      `json:"v2"`
	TagStatus       string    `json:"tag_status"`
	TagLastPulled   time.Time `json:"tag_last_pulled"`
	TagLstaPushed   time.Time `json:"tag_last_pushed"`
}

type Image struct {
	Architecture string    `json:"architecture"`
	Features     string    `json:"features"`
	Variant      string    `json:"variant"`
	Digest       string    `json:"digest"`
	OS           string    `json:"os"`
	OSFeatures   string    `json:"os_features"`
	OSVersion    string    `json:"os_version"`
	Size         int       `json:"size"`
	Status       string    `json:"status"`
	LastPulled   time.Time `json:"last_pulled"`
	PastPushed   time.Time `json:"last_pushed"`
}

type StoredImage struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
}

func main() {
	// List from Official Image wget -q -O - https://hub.docker.com/v2/namespaces/library/repositories/debian/tags?page_size=100 | jq > query.json
	// Personal repository: wget -q -O - https://hub.docker.com/v2/repositories/simonstiil/kvdb/tags?page_size=100
	// Based on https://stackoverflow.com/questions/28320134/how-can-i-list-all-tags-for-a-docker-image-on-a-remote-registry
	bytes, err := os.ReadFile("query.json")
	if err != nil {
		panic(err)
	}
	digestMap := make(map[string][]StoredImage)
	var tagList TagList
	json.Unmarshal(bytes, &tagList)
	for _, tag := range tagList.Results {
		for _, image := range tag.Images {
			if image.OS == "linux" && image.Architecture == arch {
				list := digestMap[image.Digest]
				if list == nil {
					list = []StoredImage{{Name: tag.Name, Digest: image.Digest}}
				} else {
					list = append(list, StoredImage{Name: tag.Name, Digest: image.Digest})
				}
				digestMap[image.Digest] = list
			}
		}
	}
	for digest, list := range digestMap {
		fmt.Printf("%+v : ", digest)
		for i, image := range list {
			prefix := ", "
			if i == 0 {
				prefix = ""
			}
			fmt.Printf("%v%v", prefix, image.Name)
		}
		fmt.Printf("\n")
	}
	fmt.Printf("Properties:\n")

	env := "https://raw.githubusercontent.com/SimonStiil/keyvaluedatabase/main/package.env"
	resp, err := http.Get(env)
	if err != nil {
		panic(err)
	}
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	properties, err := properties.Load(bodyText, properties.UTF8)
	if err != nil {
		panic(err)
	}
	base, baseok := properties.Get("PACKAGE_CONTAINER_BASE")
	tag, tagok := properties.Get("PACKAGE_CONTAINER_BASE_TAG")
	if baseok {
		if tagok {
			fmt.Printf("%v:%v\n", base, tag)
		} else {
			fmt.Printf("%v:%v\n", base, "latest")
		}
	}
}
