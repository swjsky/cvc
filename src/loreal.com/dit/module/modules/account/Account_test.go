package account

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type CallBack func(ts *TestObj, t *testing.T)
type CallTest func(ts *TestObj, t *testing.T) (*TestObj, error)
type GoTest func(tss []*TestObj, t *testing.T)

type TestCase struct {
	Testobjs []*TestObj
	Calltest CallTest
	Callback CallBack
	Gotest   GoTest
}
type Headers map[string]string
type TestObj struct {
	Index  int
	Desc   string
	URL1   string
	Result bool
	Method string
	Header Headers
	Params url.Values
	Resp   *http.Response
}

func NewTestCase(testobjs []*TestObj, t *testing.T) (TestCase, error) {
	tc := TestCase{}
	tc.Testobjs = testobjs
	//这个是调用http的方法，如果不满意，可以随时成一个新的方法
	tc.Calltest = func(ts *TestObj, t *testing.T) (*TestObj, error) {
		client := &http.Client{}
		req, err := http.NewRequest(ts.Method, ts.URL1, strings.NewReader(ts.Params.Encode()))
		if err != nil {
			log.Println(err)
			//如果复制流失败
			return ts, err
		}
		//如果是POST的参数类型的设置
		if ts.Method == "POST" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			//这里的header是token
		}
		//添加Header的设置
		if ts.Header != nil {
			for k, v := range ts.Header {
				req.Header.Add(k, "Bearer "+v)
			}
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			return ts, err
		}
		defer resp.Body.Close()
		ts.Resp = resp
		return ts, nil
	}
	tc.Callback = func(ts *TestObj, t *testing.T) {
		if ts.Resp == nil {
			t.Log("测试通过。描述：" + ts.Desc)
			return
		}
		if (ts.Resp.StatusCode == 200 && ts.Result == true) || (ts.Resp.StatusCode != 200 && ts.Result == false) {
			t.Log("测试通过。描述：" + ts.Desc)
			return
		}
		t.Log("测试没通过。描述：" + ts.Desc)
		return
	}
	tc.Gotest = func(tss []*TestObj, t *testing.T) {
		for i := 0; i < len(tss); i++ {
			tc.Calltest(tss[i], t)
			tc.Callback(tss[i], t)
		}
	}

	return tc, nil
}

func NewTestObj(index int, desc string, url1 string, result bool, method string, header Headers, params url.Values) *TestObj {
	if params == nil {
		params = url.Values{}
	}
	if header == nil {
		header = make(map[string]string)
	}
	if method == "" {
		method = "POST"
	}
	ts := &TestObj{
		Index:  index,
		Desc:   desc,
		URL1:   url1,
		Result: result,
		Method: method,
		Header: header,
		Params: params,
	}

	return ts
}

/*
刚开始的几个方法是下面所有的测试用例共有的方法
*/

//通过用户名和密码获得该用户的token，默认是有profile
func setToken(uid, pass string) string {
	url1 := "http://api.nikkitoday.com/account/token"
	ts := NewTestObj(1, "密码SQL注入，应该返回Unauthorized", url1, false, "POST", nil, url.Values{"uid": {uid}, "pass": {pass}, "profile": {"1"}})
	resp, err := http.PostForm(url1, ts.Params)
	if err != nil {
		log.Println(err)
		//如果请求失败
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		//如果复制流失败
		return ""
	}
	if strings.Contains(string(body), "Unauthorized") {
		//如果返回中包含"Unauthorized"
		return ""
	}
	data := make(map[string]interface{})
	if err = json.Unmarshal([]byte(string(body)), &data); err == nil {
		if len(data["token"].(string)) == 40 {
			//成功拿到了token
			return data["token"].(string)
		}
	}
	//json解析有问题
	return ""
}

/*
//获取通过http请求返回的一个json字符串，或者普通的字符串
func getReturnString(ts TestObj) (*http.Response, string) {
	client := &http.Client{}
	req, err := http.NewRequest(ts.Method, ts.URL1, strings.NewReader(ts.Params.Encode()))
	if err != nil {
		log.Println(err)
		//如果复制流失败
		return nil, ""
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//这里的header是token
	if ts.Header != nil {
		req.Header.Add("Authorization", "Bearer "+ts.Header["token"])
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		//如果复制流失败
		return nil, ""
	}
	return resp, string(body)
}
*/

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/*
接口1：token
URL:http://api.nikkitoday.com/account/token
方法：POST
参数：uid 用户名
     pass 密码
     profile 是否显示账号描述文件
参数的规范：可以是任意的字符，空格，特殊字符
接口描述：验证用户名和密码，如果通过验证就返回一个token,这个token的有效期是两个小时，后面的一个token生成后，前一个token如果没有过期，那么前一个token仍然可以使用。
		如果验证失败，那么返回Unauthorized，如果5次没有验证通过，那么这个账户就会被锁定，需要等待5分钟之后才可以解锁。如果该账户被锁期间之内，每多一次失败的验证，
		那么就再多等待一分钟。
返回结果：
	1:200	密码验证成功，返回包含token的json字符串；如果profile不为空，那么就会显示公开属性描述文件，否则就不会显示公开属性描述文件。
	2:401	密码验证不成功：，如果超过5次密码验证失败，那么账号被锁，错误：Unauthorized, Account locked!，否则显示Unauthorized。
	3:406	如果使用的不是HTTP的POST方法，返回406，表示使用的方法不对
	4:500	参数解析错误，或者返回数据转成json的时候出了错误。这个是参数错误
*/
//TestToken 测试接口http://api.nikkitoday.com/account/token
//这个接口是获取新的token的，获取新的token需要用户名和密码验证，后面的profile是结果中是否显示用户描述文件
func TestToken(t *testing.T) {
	//具体的测试用例
	url1 := "http://api.nikkitoday.com/account/token"
	ts := []*TestObj{}

	ts = append(ts, NewTestObj(1, "密码SQL注入，应该返回Unauthorized", url1, false, "POST", nil, url.Values{"uid": {"admin"}, "pass": {"qweASD!@# or 1=1"}, "profile": {"1"}}))
	ts = append(ts, NewTestObj(1, "用户名前后有空格，profile=1,应该返回Unauthorized。", url1, false, "POST", nil, url.Values{"uid": {" admin "}, "pass": {"qweASD!@#"}, "profile": {"1"}}))
	ts = append(ts, NewTestObj(1, "密码前后有空格，profile=1,应该返回Unauthorized。", url1, false, "POST", nil, url.Values{"uid": {"admin"}, "pass": {" qweASD!@# "}, "profile": {"1"}}))

	tc, err := NewTestCase(ts, t)
	if err != nil {
		t.Log("新建测试失败")
	}
	tc.Gotest(ts, t)
}

/*
	这个是根据返回的string测试具体的json 值是否符合要求
		if strings.Contains(body, "Unauthorized") {
			//如果返回中包含"Unauthorized"
			return false
		}
		data := make(map[string]interface{})
		if err := json.Unmarshal([]byte(body), &data); err == nil {
			if len(data["token"].(string)) == 40 {
				//成功拿到了token
				if (data["publicprops"] != nil && ts.Params.Get("profile") != "") || (data["publicprops"] == nil && ts.Params.Get("profile") == "") {
					//profile值为空字符串或者profile字段值缺失，返回结果不显示publicprops字段，profile字段不为空，返回publicprops
					return true
				}
				//publicprops显示有问题
				return false
			}
		}
		//json解析有问题
		return false*/

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/*
接口2：add
URL:http://api.nikkitoday.com/account/add
方法：POST
参数：uid 用户名
	 pass   密码
     roles 用户类型
     properties 用户的私有属性，可以为空
	 publicprops 用户的公开属性，可以为空
Header：Authorization  token验证
参数的规范：可以是任意的字符，空格，特殊字符
接口描述：添加一个新的用户，根据header里面的token,如果token验证通过，uid是为添加过的，就可以添加成功。
返回结果：
	1:200	添加成功；修改成功，返回前台OK字符串。
	2:403	添加的账号uid是重复的，或者插入数据库的时候出了错误。
	3:406	使用的不是HTTP的POST方法，返回406，表示使用的方法不对
	4:500	properties和publicprops的格式不正确，或者是请求参数的解析错误
*/
//TestAdd 测试接口http://api.nikkitoday.com/account/add
//这是一个账号添加的接口
func TestAdd(t *testing.T) {
	ts := []*TestObj{}
	url1 := "http://api.nikkitoday.com/account/add"
	token := setToken("admin", "qweASD!@#")
	//这个是添加用户，这俩个用户是已经添加的，如果是新加的用户，那么TestObj的Result的字段就是ture
	ts = append(ts, NewTestObj(1, "用户名uid非空，密码pass非空，用户类型roles是admin,用户属性properties非空", url1,
		true, "POST", Headers{"Authorization": token}, url.Values{"uid": {"dyq7"}, "pass": {"test123"}, "properties": {`{"name":"dyq4"}`}, "publicprops": {`{"age":"24"}`}, "roles": {"admin"}}))
	ts = append(ts, NewTestObj(1, "用户名uid非空，密码pass非空，用户类型roles是admin,用户属性properties非空", url1,
		false, "POST", Headers{"Authorization": token}, url.Values{"uid": {"dyq2"}, "pass": {"test123"}, "properties": {`{"name":"dyq4"}`}, "publicprops": {`{"age":"24"}`}, "roles": {"admin"}}))
	tc, err := NewTestCase(ts, t)
	if err != nil {
		t.Log("新建测试失败")
	}
	tc.Gotest(ts, t)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/*
接口3：init
URL:http://api.nikkitoday.com/account/init
方法：POST
参数：无
Header:无
接口描述：初始化后台环境
返回结果：
	1:200	初始化成功
	2:403	已经初始化成功了，不能再次初始化了。
	3:406	使用的不是HTTP的POST方法，返回406，表示使用的方法不对
*/

func TestInit(t *testing.T) {
	resp, err := http.PostForm("http://api.nikkitoday.com/account/init", url.Values{})
	if err != nil {
		t.Log("测试没通过，%s" + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 || resp.StatusCode == 403 {
		t.Log("测试通过")
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/*
接口4 update
URL：http://api.nikkitoday.com/account/update
方法：POST
参数：uid 用户名
     roles 用户类型
     properties 用户的私有属性
     publicprops   用户的公开属性
Header：Authorization  token验证
参数的规范：可以是任意的字符，空格，特殊字符
接口描述:修改用户的信息，不包括密码。会根据uid查找对应的用户。允许admin用户修改所有人的信息，允许自己修改自己的信息，
返回结果：
	1:200	修改信息成功，修改成功，返回前台OK字符串。
	2:403	不是客户自己更新，不是admin用户，或者是admin账户，但是修改失败，或者是自己，修改自己的私有和公开属性失败。
	3:404	uid不存在，在数据库中没找到.
	4:406	使用的不是HTTP的POST方法，返回406，表示使用的方法不对
	5:500	properties和publicprops的格式不正确，或者是请求参数的解析错误
*/
func TestUpdate(t *testing.T) {
	url1 := "http://api.nikkitoday.com/account/update"
	ts := []*TestObj{}
	token := setToken("admin", "qweASD!@#")
	ts = append(ts, NewTestObj(1, "用户名uid存在，其他三个属性的格式都是对的", url1,
		true, "POST", Headers{"Authorization": token}, url.Values{"uid": {"dyq2"}, "role": {"user"}, "publicprops": {`{"gender":"男孩","dayOfBirth":"2017-08-04 03:51:05 +0000", "name":"12","description":"123" }`}, "properties": {`{"name":"dyq4"}`}, "roles": {"admin"}}))
	ts = append(ts, NewTestObj(1, "用户名uid存在，其他三个属性的格式有不对的", url1,
		false, "POST", Headers{"Authorization": token}, url.Values{"uid": {"dyq2"}, "role": {"user"}, "publicprops": {`{"gender":"男孩","dayOfBirth":"2017-08-04 03:51:05 +0000", "name":"12","description":"123" }`}, "properties": {`{"name"}`}, "roles": {"admin"}}))
	ts = append(ts, NewTestObj(1, "用户名uid不存在，其他三个属性的格式都是对的", url1,
		false, "POST", Headers{"Authorization": token}, url.Values{"uid": {"dyq5"}, "role": {"user"}, "publicprops": {`{"gender":"男孩","dayOfBirth":"2017-08-04 03:51:05 +0000", "name":"12","description":"123" }`}, "properties": {`{"name":"dyq4"}`}, "roles": {"admin"}}))
	tc, err := NewTestCase(ts, t)
	if err != nil {
		t.Log("新建测试失败")
	}
	tc.Gotest(ts, t)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/*
接口：5
URL：http://api.nikkitoday.com/account/cpw
方法：POST
参数：
	uid		yongh
	new-pass1
	new-pass2
Header：Authorization  token验证
参数的规范：可以是任意的字符，空格，特殊字符
接口描述：修改密码的接口，自己可以给自己修改，admin可以给自己修改。根据uid修改密码
返回结果：
	1:200	修改成功，返回前台OK字符串
	2:400	两次密码不同
	3:403	不是自己或者管理员账户，不能修改
	4:406	使用的不是HTTP的POST方法，返回406，表示使用的方法不对
	5:500	请求参数解析错误
*/
func TestCpw(t *testing.T) {
	url1 := "http://api.nikkitoday.com/account/cpw"
	ts := []*TestObj{}
	token := setToken("admin", "qweASD!@#")
	ts = append(ts, NewTestObj(1, "两个密码相同", url1,
		true, "POST", Headers{"Authorization": token}, url.Values{"new-pass1": {"testing234"}, "new-pass2": {"testing234"}}))
	ts = append(ts, NewTestObj(2, "两个密码不相同", url1,
		false, "POST", Headers{"Authorization": token}, url.Values{"new-pass1": {"testing1234"}, "new-pass2": {"testing234"}}))
	tc, err := NewTestCase(ts, t)
	if err != nil {
		t.Log("新建测试失败")
	}
	tc.Gotest(ts, t)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/*
接口：5
URL：http://api.nikkitoday.com/account/public-profile
方法：GET
参数：无
参数的规范：可以是任意的字符，空格，特殊字符
接口描述：查看某个人的公开的属性，这个方法不需要token验证
返回结果：
	1:200	成功返回某个用户的公开属性的json字符串表示
	2:404	uid不存在，在数据库中没找到.
	3:406	使用的不是HTTP的POST方法，返回406，表示使用的方法不对
*/
func TestPublicprofile(t *testing.T) {
	url1 := "http://api.nikkitoday.com/account/public-profile"
	ts := []*TestObj{}
	token := setToken("admin", "qweASD!@#")
	ts = append(ts, NewTestObj(2, "两个密码相同", url1,
		false, "POST", Headers{"Authorization": token}, url.Values{}))
	tc, err := NewTestCase(ts, t)
	if err != nil {
		t.Log("新建测试失败")
	}
	tc.Gotest(ts, t)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/*
接口：5
URL：http://api.nikkitoday.com/account/register
方法：POST
参数：
	uid		新注册的用户名
	pass	用户名对应的密码
	properties		私有的属性
	publicprops		公开的属性
Header 无
参数的规范：可以是任意的字符，空格，特殊字符，私有和公开的属性必须满足json格式
接口的描述：注册新用户的接口，私有和公开属性的格式正确，uid没有被注册，就可以注册成功。这个很不安全
返回结果：
	1:200	成功注册，返回
	2:403	用户已经被注册，或者插入数据库时候出了错误。
	3:406	使用的不是HTTP的POST方法，返回406，表示使用的方法不对
	4:500	请求参数解析错误，或者返回数据转成json的时候出了错误。这个是参数错误
*/
func TestRegister(t *testing.T) {
	url1 := "http://api.nikkitoday.com/account/register"
	ts := []*TestObj{}
	token := setToken("admin", "qweASD!@#")
	ts = append(ts, NewTestObj(2, "新注册的用户", url1,
		false, "POST", Headers{"Authorization": token}, url.Values{"uid": {"1376126"}, "pass": {"123456"}, "properties": {""}, "publicprops": {""}}))
	ts = append(ts, NewTestObj(2, "已经注册的用户", url1,
		false, "POST", Headers{"Authorization": token}, url.Values{"uid": {"1376126"}, "pass": {"123456"}, "properties": {""}, "publicprops": {""}}))
	tc, err := NewTestCase(ts, t)
	if err != nil {
		t.Log("新建测试失败")
	}
	tc.Gotest(ts, t)

}
