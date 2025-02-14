package mirror
import (
    "net/url"
    "os"
    "wget/download"
)

func MirrorWebsite(websiteURL string, rateLimit string) error {
    baseURL, err := url.Parse(websiteURL)
    if err != nil {
        return err
    }

    domain := baseURL.Hostname()
    err = os.MkdirAll(domain, os.ModePerm)
    if err != nil {
        return err
    }

	err = download.DownloadFile(websiteURL, "index.html", domain, rateLimit)
    if err != nil {
        return err
    }

    // TODO: Implement recursive downloading of linked resources
    return nil
}
