package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "mime/multipart"
    "net/http"
    "os"
    "github.com/gorilla/mux"
)

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/my-webhook", receiveStackHawkWebhook).Methods("POST")

    http.Handle("/", r)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func receiveStackHawkWebhook(w http.ResponseWriter, r *http.Request) {
    var payload map[string]interface{}
    err := json.NewDecoder(r.Body).Decode(&payload)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    uploadToDefectDojo(payload)
}

func uploadToDefectDojo(payload map[string]interface{}) {
    var b bytes.Buffer
    w := multipart.NewWriter(&b)

    // Assuming payload is the data to be uploaded.
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        log.Fatal(err)
    }

    // Assuming filename is a fixed string "stackhawk-webhook.json", you might need to adjust this part
    filename := "stackhawk-webhook.json"

    part, err := w.CreateFormFile("file", filename)
    if err != nil {
        log.Fatal(err)
    }
    part.Write(payloadBytes)

    w.WriteField("scan_type", "StackHawk HawkScan")
    w.WriteField("product_name", "LibreView")
    w.WriteField("engagement_name", "ongoing remote scans")
    w.WriteField("test_title", "StackHawk")

    err = w.Close()
    if err != nil {
        log.Fatal(err)
    }

    // Using environment variable for the URL
    url := os.Getenv("DDURL") + "/api/v2/reimport-scan/"
    req, err := http.NewRequest("POST", url, &b)
    if err != nil {
        log.Fatal(err)
    }
    req.Header.Set("Content-Type", w.FormDataContentType())
    req.Header.Set("Authorization", "Token "+os.Getenv("DDAPIKEY"))

    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
        log.Fatal(err)
    }
    defer res.Body.Close()

    fmt.Println("Response: ", res)
}
