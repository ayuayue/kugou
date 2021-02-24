package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type Req struct {
	Status  int    `json:"status"`
	Error   string `json:"error"`
	Data    Data   `json:"data"`
	Errcode int    `json:"errcode"`
}
type Data struct {
	Info []Info `json:"info"`
}
type Info struct {
	Hash       string `json:"hash"`
	AlbumID    string `json:"album_id"`
	SingerName string `josn:"singername"`
	SongName   string `json:"songname"`
}
type Song struct {
	Status  int      `json:"status"`
	ErrCode int      `json:"err_code"`
	Data    SongData `json:"data"`
}
type SongData struct {
	AudioName string `json:"audio_name"`
	Lyrics    string `json:"lyrics"`
	PlayURL   string `json:"play_url"`
}

var SearchApi = "http://msearchcdn.kugou.com/api/v3/search/song"
var SongApi = "https://www.kugou.com/yy/index.php?r=play"

func query() Req {
	var name string
	flag.StringVar(&name, "name", "", "歌曲名")
	flag.Parse()
	//fmt.Println(name)
	client, err := http.NewRequest(http.MethodGet, SearchApi, nil)

	if err != nil {
		panic("请求失败")
	}
	params := make(url.Values)
	params.Add("keyword", name)
	client.URL.RawQuery = params.Encode()
	//fmt.Println(client)
	req, err := http.DefaultClient.Do(client)

	defer func() {
		req.Body.Close()
	}()
	if req.StatusCode != 200 {
		panic("返回错误响应")
	}
	body, _ := ioutil.ReadAll(req.Body)
	r := Req{}
	json.Unmarshal(body, &r)
	//fmt.Printf("%#v",r)
	return r
}

func (r *Req) List() {
	if r.Status != 1 {
		panic(r.Error)
	}
	songs := r.Data.Info
	for k, song := range songs {
		fmt.Println(k+1, song.SongName, "--", song.SingerName)
	}
}
func (r *Req) GetLink(num int) []string {
	songApi := "https://www.kugou.com/song"
	songs := r.Data.Info
	hash := songs[num-1].Hash
	album_id := songs[num-1].AlbumID
	fmt.Printf("在线播放链接：%s/#hash=%s&album_id=%s\n", songApi, hash, album_id)
	return []string{
		hash, album_id,
	}
}
func (s *Song) Download() {
	if s.Status != 1 {
		panic("获取歌曲信息失败")
	}
	lyrics := s.Data.Lyrics
	fmt.Println("开始下载歌词")
	lrcFileName := fmt.Sprintf("%s.lrc", s.Data.AudioName)
	songName := fmt.Sprintf("%s.mp3", s.Data.AudioName)
	lf, err := os.Create(lrcFileName)
	if err != nil {
		panic(err)
	}
	defer func() { _ = lf.Close() }()
	_, err = lf.Write([]byte(lyrics))
	if err != nil {
		panic(err)
	}
	fmt.Println("下载歌词完成")
	fmt.Println("开始下载歌曲")
	req, err := http.Get(s.Data.PlayURL)
	if err != nil {
		panic("请求失败")
	}
	defer func() {
		req.Body.Close()
	}()
	if req.StatusCode != 200 {
		panic("返回错误响应")
	}
	sf, err := os.Create(songName)
	if err != nil {
		panic(err)
	}
	defer func() { _ = sf.Close() }()
	_, err = io.Copy(sf, req.Body)
	if err != nil {
		panic(err)
	}
}
func main() {
	var num int
	req := query()
	req.List()
	fmt.Println("请输入要选择的歌曲链接的序号：")
	fmt.Scanln(&num)
	params := req.GetLink(num)
	song := GetSong(params)
	song.Download()
}
func GetSong(params []string) Song {
	url := fmt.Sprintf("%s/getdata&hash=%s&album_id=%s&mid=123", SongApi, params[0], params[1])
	req, err := http.Get(url)
	if err != nil {
		panic("请求失败")
	}
	defer func() {
		req.Body.Close()
	}()
	if req.StatusCode != 200 {
		panic("返回错误响应")
	}
	body, _ := ioutil.ReadAll(req.Body)
	song := Song{}
	json.Unmarshal(body, &song)
	return song
}
