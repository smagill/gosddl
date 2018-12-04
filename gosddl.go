package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

type entryACLInternal struct {
	AccountSid        string   `json:"accountsid"`
	AceType           string   `json:"aceType"`
	AceFlags          []string `json:"aceflags"`
	Rights            []string `json:"rights"`
	ObjectGuid        string   `json:"objectguid"`
	InheritObjectGuid string   `json:"InheritObjectGuid"`
}

type permissons struct {
	Owner     string             `json:"owner"`
	Primary   string             `json:"primary"`
	Dacl      []entryACLInternal `json:"dacl"`
	DaclInher []string           `json:"daclInheritFlags"`
	Sacl      []entryACLInternal `json:"sacl"`
	SaclInger []string           `json:"saclInheritFlags"`
}

var sddlRights = map[string]string{
	// Generic access rights
	"GA": "GENERIC_ALL",
	"GR": "GENERIC_READ",
	"GW": "GENERIC_WRITE",
	"GX": "GENERIC_EXECUTE",
	// Standard access rights
	"RC": "READ_CONTROL",
	"SD": "DELETE",
	"WD": "WRITE_DAC",
	"WO": "WRITE_OWNER",
	// Directory service object access rights
	"RP": "ADS_RIGHT_DS_READ_PROP",
	"WP": "ADS_RIGHT_DS_WRITE_PROP",
	"CC": "ADS_RIGHT_DS_CREATE_CHILD",
	"DC": "ADS_RIGHT_DS_DELETE_CHILD",
	"LC": "ADS_RIGHT_ACTRL_DS_LIST",
	"SW": "ADS_RIGHT_DS_SELF",
	"LO": "ADS_RIGHT_DS_LIST_OBJECT",
	"DT": "ADS_RIGHT_DS_DELETE_TREE",
	"CR": "ADS_RIGHT_DS_CONTROL_ACCESS",
	// File access rights
	"FA": "FILE_ALL_ACCESS",
	"FR": "FILE_GENERIC_READ",
	"FW": "FILE_GENERIC_WRITE",
	"FX": "FILE_GENERIC_EXECUTE",
	// Registry key access rights
	"KA": "KEY_ALL_ACCESS",
	"KR": "KEY_READ",
	"KW": "KEY_WRITE",
	"KX": "KEY_EXECUTE",
	// Mandatory label rights
	"NR": "SYSTEM_MANDATORY_LABEL_NO_READ_UP",
	"NW": "SYSTEM_MANDATORY_LABEL_NO_WRITE_UP",
	"NX": "SYSTEM_MANDATORY_LABEL_NO_EXECUTE",
}

var sddlInheritanceFlags = map[string]string{
	"P":  "DDL_PROTECTED",
	"AI": "SDDL_AUTO_INHERITED",
	"AR": "SDDL_AUTO_INHERIT_REQ",
}

var sddlAceType = map[string]string{
	"D":  "ACCESS DENIED",
	"OA": "OBJECT ACCESS ALLOWED",
	"OD": "OBJECT ACCESS DENIED",
	"AU": "SYSTEM AUDIT",
	"OU": "OBJECT SYSTEM AUDIT",
	"OL": "OBJECT SYSTEM ALARM",
	"A":  "ACCESS ALLOWED",
}

var sddlAceFlags = map[string]string{
	"CI": "CONTAINER INHERIT",
	"OI": "OBJECT INHERIT",
	"NP": "NO PROPAGATE",
	"IO": "INHERITANCE ONLY",
	"ID": "ACE IS INHERITED",
	"SA": "SUCCESSFUL ACCESS AUDIT",
	"FA": "FAILED ACCESS AUDIT",
}

var sddlSidsRep = map[string]string{
	"O":  "Owner",
	"AO": "Account operators",
	"PA": "Group Policy administrators",
	"RU": "Alias to allow previous Windows 2000",
	"IU": "Interactively logged-on user",
	"AN": "Anonymous logon",
	"LA": "Local administrator",
	"AU": "Authenticated users",
	"LG": "Local guest",
	"BA": "Built-in administrators",
	"LS": "Local service account",
	"BG": "Built-in guests",
	"SY": "Local system",
	"BO": "Backup operators",
	"NU": "Network logon user",
	"BU": "Built-in users",
	"NO": "Network configuration operators",
	"CA": "Certificate server administrators",
	"NS": "Network service account",
	"CG": "Creator group",
	"PO": "Printer operators",
	"CO": "Creator owner",
	"PS": "Personal self",
	"DA": "Domain administrators",
	"PU": "Power users",
	"DC": "Domain computers",
	"RS": "RAS servers group",
	"DD": "Domain controllers",
	"RD": "Terminal server users",
	"DG": "Domain guests",
	"RE": "Replicator",
	"DU": "Domain users",
	"RC": "Restricted code",
	"EA": "Enterprise administrators",
	"SA": "Schema administrators",
	"ED": "Enterprise domain controllers",
	"SO": "Server operators",
	"WD": "Everyone",
	"SU": "Service logon user",
}

var sddlWellKnownSidsRep = map[string]string{
	"S-1-0":        "Null Authority",
	"S-1-0-0":      "Nobody",
	"S-1-1":        "World Authority",
	"S-1-1-0":      "Everyone",
	"S-1-2":        "Local Authority",
	"S-1-2-0":      "Local ",
	"S-1-2-1":      "Console Logon ",
	"S-1-3":        "Creator Authority",
	"S-1-3-0":      "Creator Owner",
	"S-1-3-1":      "Creator Group",
	"S-1-3-2":      "Creator Owner Server",
	"S-1-3-3":      "Creator Group Server",
	"S-1-3-4":      "Owner Rights ",
	"S-1-4":        "Non-unique Authority",
	"S-1-5":        "NT Authority",
	"S-1-5-1":      "Dialup",
	"S-1-5-2":      "Network",
	"S-1-5-3":      "Batch",
	"S-1-5-4":      "Interactive",
	"S-1-5-6":      "Service",
	"S-1-5-7":      "Anonymous",
	"S-1-5-8":      "Proxy",
	"S-1-5-9":      "Enterprise Domain Controllers",
	"S-1-5-10":     "Principal Self",
	"S-1-5-11":     "Authenticated Users",
	"S-1-5-12":     "Restricted Code",
	"S-1-5-13":     "Terminal Server Users",
	"S-1-5-14":     "Remote Interactive Logon ",
	"S-1-5-15":     "This Organization ",
	"S-1-5-17":     "This Organization ",
	"S-1-5-18":     "Local System",
	"S-1-5-19":     "NT Authority",
	"S-1-5-20":     "NT Authority",
	"S-1-5-32-544": "Administrators",
	"S-1-5-32-545": "Users",
	"S-1-5-32-546": "Guests",
	"S-1-5-32-547": "Power Users",
	"S-1-5-32-548": "Account Operators",
	"S-1-5-32-549": "Server Operators",
	"S-1-5-32-550": "Print Operators",
	"S-1-5-32-551": "Backup Operators",
	"S-1-5-32-552": "Replicators",
	"S-1-5-64-10":  "NTLM Authentication ",
	"S-1-5-64-14":  "SChannel Authentication ",
	"S-1-5-64-21":  "Digest Authentication ",
	"S-1-5-80":     "NT Service ",
	"S-1-5-80-0":   "All Services ",
	"S-1-5-83-0":   "NT VIRTUAL MACHINE\\Virtual Machines",
	"S-1-16-0":     "Untrusted Mandatory Level ",
	"S-1-16-4096":  "Low Mandatory Level ",
	"S-1-16-8192":  "Medium Mandatory Level ",
	"S-1-16-8448":  "Medium Plus Mandatory Level ",
	"S-1-16-12288": "High Mandatory Level ",
	"S-1-16-16384": "System Mandatory Level ",
	"S-1-16-20480": "Protected Process Mandatory Level ",
	"S-1-16-28672": "Secure Process Mandatory Level ",
	"S-1-5-32-554": "BUILTIN\\Pre-Windows 2000 Compatible Access",
	"S-1-5-32-555": "BUILTIN\\Remote Desktop Users",
	"S-1-5-32-556": "BUILTIN\\Network Configuration Operators",
	"S-1-5-32-557": "BUILTIN\\Incoming Forest Trust Builders",
	"S-1-5-32-558": "BUILTIN\\Performance Monitor Users",
	"S-1-5-32-559": "BUILTIN\\Performance Log Users",
	"S-1-5-32-560": "BUILTIN\\Windows Authorization Access Group",
	"S-1-5-32-561": "BUILTIN\\Terminal Server License Servers",
	"S-1-5-32-562": "BUILTIN\\Distributed COM Users",
	"S-1-5-32-569": "BUILTIN\\Cryptographic Operators",
	"S-1-5-32-573": "BUILTIN\\Event Log Readers ",
	"S-1-5-32-574": "BUILTIN\\Certificate Service DCOM Access ",
	"S-1-5-32-575": "BUILTIN\\RDS Remote Access Servers",
	"S-1-5-32-576": "BUILTIN\\RDS Endpoint Servers",
	"S-1-5-32-577": "BUILTIN\\RDS Management Servers",
	"S-1-5-32-578": "BUILTIN\\Hyper-V Administrators",
	"S-1-5-32-579": "BUILTIN\\Access Control Assistance Operators",
	"S-1-5-32-580": "BUILTIN\\Remote Management Users",
	"S-1-5-80-956008885-3418522649-1831038044-1853292631-2271478464": "Trusted Installer",
}

func sidReplace(str string) string {
	// replace identification account: sid/wellkhownsid/usersid
	if len(str) > 2 {
		if x, ok := sddlWellKnownSidsRep[str]; ok {
			return x
		} else {
			return str
		}
		return replacer(sddlWellKnownSidsRep, str)[0]
	} else {
		return replacer(sddlSidsRep, str)[0]
	}
}

func replacer(maps map[string]string, str string) []string {
	// Chunk string with 2 letters, add to array and then resolve
	var temp, result []string
	if len(str) > 2 {
		for j := 0; j < len(str)-1; j = j + 2 {
			temp = append(temp, fmt.Sprintf("%s%s", string(str[j]), string(str[j+1])))
		}
	} else {
		temp = append(temp, str)
	}
	for _, v := range temp {
		if x, ok := maps[v]; ok {
			result = append(result, x)
		} else {
			result = append(result, v)
		}
	}
	return result
}

func GetInfo(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("Hello")
}

func splitBodyACL(str string) entryACLInternal {
	// Base format ACL: (ace_type;ace_flags;rights;object_guid;inherit_object_guid;account_sid)
	// Convert values from string to struct with replace strings
	temp := strings.Split(str, ";")
	return entryACLInternal{
		AceType:           replacer(sddlAceType, temp[0])[0],
		AceFlags:          replacer(sddlAceFlags, temp[1]),
		Rights:            replacer(sddlRights, temp[2]),
		ObjectGuid:        temp[3],
		InheritObjectGuid: temp[4],
		AccountSid:        sidReplace(temp[5]),
	}
}

func splitBody(body string) []entryACLInternal {
	var entryACLInternalArr []entryACLInternal
	for _, y := range strings.Split(body, "(") {
		if y != "" {
			ace := strings.TrimSuffix(y, ")")
			entryACLInternalArr = append(entryACLInternalArr, splitBodyACL(ace))
		}
	}
	return entryACLInternalArr
}

func (p *permissons) parseBody(body string) ([]string, []entryACLInternal) {
	var inheritFlagArr []string
	var entryACLInternalArr []entryACLInternal
	if strings.Index(body, "(") != 0 {
		inheritFlag := body[0:strings.Index(body, "(")]
		ace := body[strings.Index(body, "("):]
		if len(inheritFlag) > 2 {
			for j := 0; j < len(inheritFlag)-1; j = j + 2 {
				inheritFlagArr = append(inheritFlagArr, replacer(sddlInheritanceFlags, fmt.Sprintf("%s%s", string(inheritFlag[j]), string(inheritFlag[j+1])))[0])
			}
		}
		entryACLInternalArr = splitBody(ace)
	} else {
		entryACLInternalArr = splitBody(body)
	}
	return inheritFlagArr, entryACLInternalArr
}

func (p *permissons) parseSDDL(sddrArr []string) {
	for _, y := range sddrArr {
		sddlSplit := strings.Split(y, ":")
		letter := sddlSplit[0]
		body := sddlSplit[1]
		switch letter {
		case "O":
			p.Owner = sidReplace(body)
		case "G":
			p.Primary = sidReplace(body)
		case "D":
			p.DaclInher, p.Dacl = p.parseBody(body)
		case "S":
			p.SaclInger, p.Sacl = p.parseBody(body)
		default:
			log.Fatal("Unresolved group")
		}
	}

}

func (p *permissons) sliceSDDL(indecs []int, str string) {
	// create slice objects from str to array of strings
	var sddlArr []string
	for i := 0; i < len(indecs)-1; i++ {
		sl := str[indecs[i]:indecs[i+1]]
		sddlArr = append(sddlArr, sl)
	}
	p.parseSDDL(sddlArr)
}

func (p *permissons) findIndex(str string) {
	groups := []string{"O:", "G:", "D:", "S:"}
	var result []int
	for _, i := range groups {
		if strings.Index(str, i) != -1 {
			result = append(result, strings.Index(str, i))
		}
	}
	result = append(result, len(str))
	p.sliceSDDL(result, str)
}

func Decode(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if params["sddl"] != "" {
		sddl := params["sddl"]
		var permisson permissons
		permisson.findIndex(sddl)
		json.NewEncoder(w).Encode(permisson)
		return
	}

}

func api(port string) {
	port = ":" + port
	router := mux.NewRouter()
	router.HandleFunc("/sddl", GetInfo).Methods("GET")
	router.HandleFunc("/sddl/{sddl}", Decode).Methods("GET")
	log.Fatal(http.ListenAndServe(port, router))
}

func main() {
	apiPtr := flag.Bool("api", false, "a bool")
	apiPortPtr := flag.String("port", "8000", "Default port 8000")
	flag.Parse()
	if *apiPtr {
		fmt.Println("API Interface started on port", *apiPortPtr)
		api(*apiPortPtr)
	} else if flag.Args() != nil {
		var permisson permissons
		permisson.findIndex(flag.Args()[0])
		b, err := json.Marshal(permisson)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}
}
