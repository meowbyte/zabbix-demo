package cam_monitoring

import (
	"os"
	"fmt"
	"time"
	"bytes"
	"errors"
	"strconv"
	"net/http"
	"io/ioutil"
	"zabbix.com/pkg/plugin"
	"golift.io/ffmpeg"
)

type Plugin struct {
    plugin.Base
    Time int64
}
var impl Plugin

func init() {
    plugin.RegisterMetrics(&impl, "Camera", "cam_monitoring", "Test camera.")
}

func make_req(url string, json_string string) (result []byte, err error) {
    var json = []byte(json_string)
    req, err := http.NewRequest("GET", url, bytes.NewBuffer(json))
    req.Header.Set("Content-Type", "application/json")

    text, err := http.DefaultClient.Do(req)
    if err != nil {
    	return nil, err
    }

	b, err := ioutil.ReadAll(text.Body)
	defer text.Body.Close()

	return b, nil
}

func check_stream(addr string) (result int64, err error) {
	d := time.Now()
	title := "dump"
	output := fmt.Sprintf("/tmp/%s_%s", title, d.Format("01-02-2006 15:04:05"))

	c := &ffmpeg.Config{
		FFMPEG: "/usr/bin/ffmpeg",
		Copy:   true,
		Audio:  false,
		Time:   3,	// 3 sec
	}

	encode := ffmpeg.Get(c)
	_, _, err = encode.SaveVideo(addr, output, output)

	if err != nil {
		return 0, err
	}

	f, err := os.Stat(output)
	if err != nil {
	    return 0, err
	}
	
	return f.Size(), nil
}

func (p *Plugin) Export(key string, params []string, ctx plugin.ContextProvider) (result interface{}, err error) {
    if len(params) != 1 {
        return nil, errors.New("Wrong parameters.")
    }

    //addr := "rtsp://admin:smartspaces2019@192.168.1.168:554" // real cam - not working
    addr := "rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov" // test

	size, err := check_stream(addr)
	if err != nil {
		return nil, errors.New("Camera is not available.")
	}

	if size < 1 {
		return nil, errors.New("Camera is available, but no data is received.")
	}

	now, err := strconv.ParseInt(params[0], 10, 64)
	if err != nil {
		return nil, err
	}


	return fmt.Sprintf("Camera is available for %d minutes, last dump size is %d bytes.", 
		 (time.Now().Unix() - now) / 60, size), nil
}