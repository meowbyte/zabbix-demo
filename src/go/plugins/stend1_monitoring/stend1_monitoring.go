package stend1_monitoring

import (
	"fmt"
	"bytes"
	"errors"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"zabbix.com/pkg/plugin"
)

type Value struct {
	Value float64 `json:"value"`
}

type Response struct {
	Result string `json:"result"`
	Data []Value `json:"data"`
}

type Plugin struct {
    plugin.Base
}
var impl Plugin

func init() {
    plugin.RegisterMetrics(&impl, "Monitoring", "stend1_monitoring.check_service", "Checks service on stend 1.")
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

func auth() (result interface{}, err error) {
	json_string := `{
	    "jsonrpc": "2.0",
	    "method": "user.login",
	    "params": {
	        "user": "Admin",
	        "password": "zabbix"
	    },
	    "id": 1,
	    "auth": null
	}`

	var res Response

	b, err := make_req("http://localhost/zabbix/api_jsonrpc.php", json_string)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}

	return res.Result, nil
}

func init_graphics(name string) (err error) {
	a, err := auth()

	if err != nil {
		return err
	}

	json_string := fmt.Sprintf(`{
	    "jsonrpc": "2.0",
	    "method": "graph.create",
	    "params": {
	        "name": "%s",
	        "width": 900,
	        "height": 200,
	        "gitems": [
	            {
	                "itemid": "<id>",
	                "color": "00AA00",
	                "sortorder": "0"
	            },
	        ]
	    },
	    "auth": "%s",
	    "id": 1
	}`, name, a)

	_, err = make_req("http://localhost/zabbix/api_jsonrpc.php", json_string)

	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) Export(key string, params []string, ctx plugin.ContextProvider) (result interface{}, err error) {
    if len(params) != 1 {
        return 0, errors.New("Wrong parameters.")
    }

	var res Response
	b, err := make_req(
		fmt.Sprintf("http://iot-stend1.goip.de/api/app/get_service_data?service_id=%s", 
    	params[0]), `{}`)

	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}

	length := len(res.Data)
	value := res.Data[length - 1] // last data

	/*err = init_graphics("Sensor")

	if err != nil {
		return nil, err
	}*/

	return value.Value, nil
}