package oauth2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tarent/loginsrv/model"
)

var PCAPI = "https://api.planningcenteronline.com/people/v2"

func init() {
	RegisterProvider(providerPlanningCenter)
}

// PCUser is used for parsing the PCO api response
type PCUser struct {
	Data struct {
		Id string `json:"id"`
		Attributes struct {
			Name string `json:"name"`
			People_permissions string `json:"people_permissions"`
			
		} `json:"attributes"`	
	} `json:"data"`
}

// PCOrg is used for parsing the PCO api organization info response
type PCOrg struct {
	Data struct {
		Id string `json:"id"`
		Attributes struct {
			Name string `json:"name"`	
		} `json:"attributes"`	
	} `json:"data"`
}


var providerPlanningCenter = Provider{
	Name:     "planningcenter",
	AuthURL:  "https://api.planningcenteronline.com/oauth/authorize",
	TokenURL: "https://api.planningcenteronline.com/oauth/token",
	GetUserInfo: func(token TokenInfo) (model.UserInfo, string, error) {
		
		
		/////// GET USER INFO /////////
		pcu := PCUser{}
		
		// url for getting user data after auth
		url := fmt.Sprintf("%v/me", PCAPI)
		
		// use authentication token to request the "me" page
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		auth := fmt.Sprintf("Bearer %v", token.AccessToken)
		req.Header.Add("Authorization", auth)
			
		// check for erros in the user info json
		respUser, err := client.Do(req)	
		if err != nil {
			return model.UserInfo{}, "", err
		}
		defer respUser.Body.Close()
		
		if !strings.Contains(respUser.Header.Get("Content-Type"), "json") {
			return model.UserInfo{}, "", fmt.Errorf("wrong content-type on planningcenter get user info: %v", respUser.Header.Get("Content-Type"))
		}

		if respUser.StatusCode != 200 {
			return model.UserInfo{}, "", fmt.Errorf("got http status %v on planningcenter get user info", respUser.StatusCode)
		}

		b, err := ioutil.ReadAll(respUser.Body)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error reading planningcenter get user info: %v", err)
		}
		//fmt.Println(b)


		// parse the user info json into a structure
		err = json.Unmarshal(b, &pcu)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error parsing planningcenter get user info: %v", err)
		}
		
		//fmt.Println(pcu)

		groups := make([]string, 1)

		// check if user has people_permissions, if so get the org info
		// otherwise set groups to empty
		if pcu.Data.Attributes.People_permissions == "" {
			//fmt.Println("no people permissions")
			groups = make([]string, 0) // set groups to empty
		} else {
		
			//////// GET ORG INFO  ///////

			// url for getting org data after auth
			url = PCAPI
			pcOr := PCOrg{}
			// use authentication token to request the org info page
			client = &http.Client{}
			req, err = http.NewRequest("GET", url, nil)
			auth = fmt.Sprintf("Bearer %v", token.AccessToken)
			req.Header.Add("Authorization", auth)
				
			// check for erros in the org info json
			respOrg, err := client.Do(req)	
			if err != nil {
				return model.UserInfo{}, "", err
			}
			defer respOrg.Body.Close()
			
			if !strings.Contains(respOrg.Header.Get("Content-Type"), "json") {
				return model.UserInfo{}, "", fmt.Errorf("wrong content-type on planningcenter get org info: %v", respOrg.Header.Get("Content-Type"))
			}

			if respOrg.StatusCode != 200 {
				return model.UserInfo{}, "", fmt.Errorf("got http status %v on planningcenter get org info", respOrg.StatusCode)
			}

			b, err = ioutil.ReadAll(respOrg.Body)
			if err != nil {
				return model.UserInfo{}, "", fmt.Errorf("error reading planningcenter get org info: %v", err)
			}
			//fmt.Println(b)

			// parse the org info json into a structure
			err = json.Unmarshal(b, &pcOr)
			if err != nil {
				return model.UserInfo{}, "", fmt.Errorf("error parsing planningcenter get org info: %v", err)
			}

			
			//fmt.Println(pcOr)

			// use the groups attribute to store the org id and permissions
			// only users with the right org id and people_permissions level can access
			groups[0] = fmt.Sprintf("%v|%v", pcOr.Data.Id, pcu.Data.Attributes.People_permissions)

			//fmt.Println(groups)

		}

		return model.UserInfo{
			Sub:     pcu.Data.Id,
		//	Picture: pcu.AvatarURL, // not implemented
			Name:    pcu.Data.Attributes.Name,
		//	Email:   pcu.Email,		// not implemented
			Origin:  "planningcenter",
			Groups:  groups,
		}, string(b), nil
	},
}
