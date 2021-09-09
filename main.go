package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	remote "github.com/yoyofxteam/nacos-viper-remote"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type contentTable struct {
	id          string
	ty          int `db:"type"`
	content     string
	modelMenuId int `db:"model_menu_id"`
	deleted     int
}


var (
	defaultConfig viper.Viper
)

func getPicture(table *contentTable) []string {

	pictures := make([]string, 0, 10)
	if strings.Contains(table.content, "![]") {
		//the content contain picture
		rexp := regexp.MustCompile("!\\[\\]\\((.*?)\\)")
		matches := rexp.FindAllStringSubmatch(table.content, -1)
		for _, s := range matches {

			//fmt.Println("find picture ====>", s[0])
			buffer := strings.Split(s[0], "/")
			pictureName := buffer[len(buffer)-1]
			//fmt.Println("pictrue name ====>", pictureName[0: len(pictureName) - 1])
			pictures = append(pictures, pictureName[0:len(pictureName)-1])

		}
		/*str := ""
		matched, err := regexp.MatchString("/!\\[\\]\\((.*?)\\)/g", str)
		fmt.Println(matched, err)*/

	} else {
		return nil
	}
	return pictures
}

func main() {
	var nacosAddr string
	var path string
	//if have args use args or use env
	//--nacosAddr=172.22.1.7:30501 --imgPath=D:/opt/img
	args := os.Args
	for _, arg := range args {
		tmpParam := strings.Split(arg, "=")
		if tmpParam[0] == "--nacosAddr" {
			nacosAddr = tmpParam[1]
		}
		if tmpParam[0] == "--imgPath" {
			path = tmpParam[1]
		}
	}

	//get configuration
	if nacosAddr == "" {
		nacosAddr = os.Getenv("NACOS_SERVER")
	}
	if path == "" {
		path = "/opt/img"
	}

	fmt.Println("========== begin drop useless img==========")
	runtimeViper := viper.New()

	fmt.Println("nacos address ====> {} ", nacosAddr)
	result := strings.Split(nacosAddr, ":")
	fmt.Println(result[0], "port=======>", result[1])
	port, _ := strconv.ParseUint(result[1], 10, 64)
	remote.SetOptions(&remote.Option{
		Url:         result[0],
		Port:        port,
		NamespaceId: "dev",
		GroupName:   "DEFAULT_GROUP",
		Config:      remote.Config{DataId: "sugoncloud-common-environment.yaml"},
		Auth:        nil,
	})
	err := runtimeViper.AddRemoteProvider("nacos", nacosAddr, "")
	runtimeViper.SetConfigType("yaml")
	err = runtimeViper.ReadRemoteConfig()
	_ = runtimeViper.WatchRemoteConfigOnChannel()
	var mysqlAddr string
	if err == nil {
		fmt.Println("not error")
		mysqlIp := runtimeViper.Get("mysql.ip")
		mysqlPort := runtimeViper.Get("mysql.port")
		fmt.Println("ip ======>", mysqlIp)
		fmt.Println("port  ======>", mysqlPort)
		str := fmt.Sprintf("%s", mysqlIp)
		str2 := fmt.Sprintf("%d", mysqlPort)
		mysqlAddr = str + ":" + str2
	}
	//connect db
	db, err := sql.Open("mysql", "root:password@tcp("+mysqlAddr+")/sugoncloud_guide?charset=utf8")
	if err != nil {
		fmt.Println(err)
		return
	}
	rows, _ := db.Query("select * from content")

	fmt.Println(rows)
	//get picture id
	type void struct{}
	var pictureIds map[string]void
	pictureIds = make(map[string]void)




	for rows.Next() {
		var s contentTable
		err = rows.Scan(&s.id, &s.ty, &s.content, &s.modelMenuId, &s.deleted)
		//fmt.Println(s.content)
		var tem void
		results := getPicture(&s)
		if len(results) > 0 {
			for pic := range results {
				pictureIds[results[pic]] = tem
			}
		}

	}
	rows.Close()
	//check and drop useless pictures

	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("============== check files =============")
	for _, file := range files {
		_, ok := pictureIds[file.Name()]
		if ok {
			//fmt.Println("reserve =====>", pic)
		} else {
			fmt.Println("will remove========>", path+"/"+file.Name())
			os.Remove(path+"/"+file.Name())
		}

	}

}
