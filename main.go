package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/nicholas-fedor/shoutrrr"
)

const (
	baseURL  = "https://qis.w-hs.de/qisserver/rds"
	hashFile = "/data/last_hash.txt"
	interval = 30 * time.Minute
)

var (
	username   = os.Getenv("QIS_USERNAME")
	password   = os.Getenv("QIS_PASSWORD")
	webhookURL = os.Getenv("WEBHOOK_URL")
)

func main() {
	checkGrades()
	for range time.Tick(interval) {
		checkGrades()
	}
}

func checkGrades() {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// 1. Startseite
	resp, err := client.Get(baseURL + "?state=user&type=0")
	if err != nil {
		fmt.Println("Fehler:", err)
		return
	}
	resp.Body.Close()

	// Session-ID aus Cookie
	sessionID := ""
	u, _ := url.Parse(baseURL)
	for _, cookie := range jar.Cookies(u) {
		if cookie.Name == "JSESSIONID" {
			sessionID = cookie.Value
		}
	}

	if sessionID == "" {
		sessionID = extractSessionID(resp.Request.URL.String())
	}

	if sessionID == "" {
		fmt.Println("Keine Session-ID gefunden")
		return
	}
	fmt.Println("Session-ID:", sessionID)

	// 2. Login
	loginURL := fmt.Sprintf("%s;jsessionid=%s?state=user&type=1&category=auth.login&startpage=portal.vm&breadCrumbSource=portal", baseURL, sessionID)

	form := url.Values{
		"asdf":   {username},
		"fdsa":   {password},
		"submit": {"Anmelden"},
	}

	req, _ := http.NewRequest("POST", loginURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Login fehlgeschlagen:", err)
		return
	}

	// 3. asi-Token aus der Antwort extrahieren
	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()

	asi := extractASI(doc)
	if asi == "" {
		fmt.Println("asi-Token nicht gefunden – Login vermutlich fehlgeschlagen")
		// Debug: HTML ausgeben
		html, _ := doc.Html()
		fmt.Println(html[:min(500, len(html))])
		return
	}
	fmt.Println("asi-Token:", asi)

	// 4. Notenseite abrufen
	gradesURL := fmt.Sprintf("%s?state=notenspiegelStudent&next=list.vm&nextdir=qispos/notenspiegel/student&createInfos=Y&struct=auswahlBaum&nodeID=%s&expand=0&asi=%s",
		baseURL,
		url.QueryEscape("auswahlBaum|abschluss:abschl=53,stgnr=1|studiengang:stg=888"),
		asi,
	)

	resp, err = client.Get(gradesURL)
	if err != nil {
		fmt.Println("Notenseite fehlgeschlagen:", err)
		return
	}
	defer resp.Body.Close()

	// 5. Parsen
	doc, _ = goquery.NewDocumentFromReader(resp.Body)
	content, _ := doc.Find("table").Html()

	// Debug: Inhalt anzeigen
	// fmt.Println(content)

	// 6. Hash check
	hash := md5.Sum([]byte(content))
	newHash := hex.EncodeToString(hash[:])
	oldHash, _ := os.ReadFile(hashFile)

	if newHash != strings.TrimSpace(string(oldHash)) {
		fmt.Println("🎉 Änderung erkannt!", time.Now().Format("15:04:05"))
		os.WriteFile(hashFile, []byte(newHash), 0644)

		err := shoutrrr.Send(webhookURL, "Neue Noten verfügbar!")
		if err != nil {
			fmt.Println("Error sending notification:", err)
		}
	} else {
		fmt.Println("Keine Änderung -", time.Now().Format("15:04:05"))
	}
}

func extractSessionID(urlStr string) string {
	re := regexp.MustCompile(`jsessionid=([A-Z0-9]+)`)
	matches := re.FindStringSubmatch(urlStr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractASI(doc *goquery.Document) string {
	// asi ist oft in Links auf der Seite
	asi := ""
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.Contains(href, "asi=") {
			re := regexp.MustCompile(`asi=([A-Za-z0-9._-]+)`)
			matches := re.FindStringSubmatch(href)
			if len(matches) > 1 {
				asi = matches[1]
			}
		}
	})
	return asi
}
