package main
import (
    "flag"
    "fmt"
    "io"
    "net"
    "net/url"
    "os"
    "strings"
)

/*
Application Work plan:

1. Take input from all sources
    a. stdin
    b. flag
    c. CGI Envs
2. Parse URL
3. output field of request
*/

/******************************************************************************/
// MARK - Structures

type AppData struct {
    Url *string
    Param *string
    Path *string
    Part *string
    Verbose *bool
    Cgi *bool
}

/******************************************************************************/
// MARK - Functions

func GetEnv(name, backup string) string {
    if value := os.Getenv(name); len(value)>0 {
        return value
    }
    return backup
}

func readFromEnv() string {
    //method := os.Getenv("REQUEST_METHOD") // GET
    scheme := GetEnv("REQUEST_SCHEME", "http") // https
    server := GetEnv("SERVER_NAME", "localhost") // thomascherry.name
    port := GetEnv("SERVER_PORT", "80") // 443
    script := GetEnv("SCRIPT_NAME", "") // /cgi-bin/all.cgi
    query := GetEnv("QUERY_STRING", "") // a=b&1=2

    full := fmt.Sprintf("%s://%s:%s%s?%s", scheme, server, port, script, query)
    return full
}

func readFromStdIn() string {
    ret := ""
    stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		everything, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		ret = string(everything)
	}
	return ret
}

func removeRedundency(raw string) string {
    //remove port when it can be assumed
    if strings.Contains(raw, "http://") &&
        (strings.Contains(raw, ":80/") || strings.Contains(raw, ":80?")) {
        raw = strings.Replace(raw, ":80", "", 1)
    } else if strings.Contains(raw, "https://") &&
        (strings.Contains(raw, ":443/") || strings.Contains(raw, ":443?")) {
        raw = strings.Replace(raw, ":443", "", 1)
    }
    //remove trailing question mark
    l := len(raw) - 1
    if raw[l] == '?' {
        raw = raw[0:l]
    }
    return raw
}

func work(sess AppData, details string) string {
    details = strings.Trim(details, " \n\t")
    //s := "postgres://user:pass@host.com:5432/path?k=v#f"
    //s = "?k=v#f"
    u, err := url.Parse(details)
    if err != nil {
        panic(err)
    }

    if *sess.Param != "" {
        m, _ := url.ParseQuery(u.RawQuery)
        list := m[*sess.Param]
        if len(list)>0 {
            return list[0]
        }
        return ""
    }

    switch *sess.Part {
    case "scheme":
        return u.Scheme
    case "user":
        return fmt.Sprintf("%+v", u.User)
    case "username":
        return u.User.Username()
    case "password":
        p, _ := u.User.Password()
        return p
    case "host-port":
        return u.Host
    case "host":
        host, _, _ := net.SplitHostPort(u.Host)
        return host
    case "port":
        _, port, _ := net.SplitHostPort(u.Host)
        return port
    case "path":
        return u.Path
    case "fragment":
        return u.Fragment
    case "query":
        m, _ := url.ParseQuery(u.RawQuery)
        return fmt.Sprintf("%s\t%+v", u.RawQuery, m)
    case "full":
        approx := fmt.Sprintf("%s://%s%s?%s", u.Scheme, u.Host, u.Path, u.RawQuery)
        approx = removeRedundency(approx)
        if approx[len(approx)-1] == '?' {
            approx = approx[0:len(approx)-1]
        }
        return approx
    }
    return ""
}

/******************************************************************************/
// MARK - Command Line Functions

/** write results to the console, support a verbose mode */
func output(sess AppData, text, who string) {
    if *sess.Verbose {
        fmt.Printf("%s\t%s\n", text, who)
    } else {
        fmt.Printf("%s\n", text)
    }
}

/**
Process the Command Line Flags into an AppData structure to be used as the app
session data
*/
func process_flags() AppData {
    session := AppData{}
    session.Url = flag.String("url", "", "Input URL to process")
    session.Verbose = flag.Bool("verbose", false, "Verbose Mode")
    session.Cgi = flag.Bool("cgi", false, "Verbose Mode")

    session.Param = flag.String("param", "", "Output Query parameter")
    session.Part = flag.String("part", "", "Output One section: scheme, user, username, password, host-port, host,port, path, fragment, query, full")

    flag.Parse()

    return session
}

func main() {
    sess := process_flags()

    if url := *sess.Url; 0<len(url) {
        output(sess, work(sess, url), "flag")
    }

    if url := readFromStdIn(); 0<len(url) {
        output(sess, work(sess, url), "stdin")
    }

    if *sess.Cgi {
        if url := readFromEnv(); 0<len(url) {
            output(sess, work(sess, url), "cgi")
        }
    }
}
