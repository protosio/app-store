package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/docker/api/types"
	"github.com/protosio/app-store/installer"
	"github.com/protosio/app-store/util"
)

var log = util.GetLogger()

type Target struct {
	MediaType  string `json:"mediaType"`
	Size       int    `json:"size"`
	Digest     string `json:"digest"`
	Length     int    `json:"length"`
	Repository string `json:"repository"`
	URL        string `json:"url"`
	Tag        string `json:"tag"`
}

type Event struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Target    Target    `json:"target"`
}

type Events struct {
	Events []Event `json:"events"`
}

func parsePublicPorts(publicports string) []util.Port {
	ports := []util.Port{}
	for _, portstr := range strings.Split(publicports, ",") {
		portParts := strings.Split(portstr, "/")
		if len(portParts) != 2 {
			log.Errorf("Error parsing installer port string %s", portstr)
			continue
		}
		portNr, err := strconv.Atoi(portParts[0])
		if err != nil {
			log.Errorf("Error parsing installer port string %s", portstr)
			continue
		}
		if portNr < 1 || portNr > 0xffff {
			log.Errorf("Installer port is out of range %s (valid range is 1-65535)", portstr)
			continue
		}
		port := util.Port{Nr: portNr}
		if strings.ToUpper(portParts[1]) == string(util.TCP) {
			port.Type = util.TCP
		} else if strings.ToUpper(portParts[1]) == string(util.UDP) {
			port.Type = util.UDP
		} else {
			log.Errorf("Invalid protocol(%s) for port(%s)", portParts[1], portParts[0])
			continue
		}
		ports = append(ports, port)
	}
	return ports
}

// parseMetadata parses the image metadata from the image labels
func parseMetadata(labels map[string]string) (installer.InstallerMetadata, error) {
	r := regexp.MustCompile("(^protos.installer.metadata.)(\\w+)")
	metadata := installer.InstallerMetadata{}
	for label, value := range labels {
		labelParts := r.FindStringSubmatch(label)
		if len(labelParts) == 3 {
			switch labelParts[2] {
			case "capabilities":
				capabilities := strings.Split(value, ",")
				caps := []map[string]string{}
				for _, cap := range capabilities {
					caps = append(caps, map[string]string{"Name": cap})
				}
				metadata.Capabilities = caps
			case "params":
				metadata.Params = strings.Split(value, ",")
			case "provides":
				metadata.Provides = strings.Split(value, ",")
			case "requires":
				metadata.Requires = strings.Split(value, ",")
			case "publicports":
				metadata.PublicPorts = parsePublicPorts(value)
			case "description":
				metadata.Description = value
			}
		}

	}
	if metadata.Description == "" {
		return metadata, errors.New("installer metadata field 'description' is mandatory")
	}
	return metadata, nil
}

func getImageTags(name string) ([]string, error) {
	var tagList struct{ Tags []string }
	log.Infof("Retrieving tags for Docker image %s", name)
	r, err := http.Get(fmt.Sprintf("http://docker-registry:5000/v2/%s/tags/list", name))
	if err != nil {
		return tagList.Tags, err
	}
	defer r.Body.Close()

	bodyJSON, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return tagList.Tags, err
	}

	err = json.Unmarshal(bodyJSON, &tagList)
	if err != nil {
		return tagList.Tags, err
	}

	return tagList.Tags, nil
}

func getImageMetadata(name string, tag string) (installer.InstallerMetadata, error) {
	var metadata installer.InstallerMetadata
	// this nasty sleep is required (for now) because the push even is triggered before the metadata is available for an image
	time.Sleep(5 * time.Second)
	log.Infof("Retrieving metadata for image %s:%s", name, tag)
	httpClient := &http.Client{}

	// Retrieves the v2 manifest for the Docker image, based on the tag. From that we extract the image/tag digest.
	url := fmt.Sprintf("http://docker-registry:5000/v2/%s/manifests/%s", name, tag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return metadata, errors.Wrap(err, "Failed to retrieve manifest")
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	r, err := httpClient.Do(req)
	if err != nil {
		return metadata, errors.Wrap(err, "Failed to retrieve manifest")
	}
	defer r.Body.Close()

	imageDigest := r.Header.Get("docker-content-digest")
	if imageDigest == "" {
		return metadata, errors.New("The image digest is empty")
	}

	bodyJSON, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return metadata, errors.Wrap(err, "Failed to read manifest body")
	}

	var manifest schema2.Manifest
	err = json.Unmarshal(bodyJSON, &manifest)
	if err != nil {
		return metadata, errors.Wrap(err, "Error unmarshaling image manifest")
	}

	// Retrieves the image inspect data which contains the installer metadata
	url = fmt.Sprintf("http://docker-registry:5000/v2/%s/blobs/%s", name, manifest.Config.Digest.String())
	r, err = http.Get(url)
	if err != nil {
		return metadata, errors.Wrap(err, "Error retrieving image blob")
	}
	defer r.Body.Close()

	bodyJSON, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return metadata, errors.Wrap(err, "Error reading image inspect data")
	}
	var imageInfo types.ImageInspect
	err = json.Unmarshal(bodyJSON, &imageInfo)
	if err != nil {
		return metadata, errors.Wrap(err, "Error unmarshalling image inspect data")
	}
	metadata, err = parseMetadata(imageInfo.Config.Labels)
	if err != nil {
		return metadata, errors.Wrap(err, "Could not parse metadata for image")
	}
	metadata.PlatformID = name + "@" + imageDigest

	return metadata, nil
}

// FullScan does a full scan of all the images in the registry and imports them
func FullScan() error {
	log.Info("Performing full Docker registry scan")
	r, err := http.Get("http://docker-registry:5000/v2/_catalog")
	if err != nil {
		return err
	}
	defer r.Body.Close()

	bodyJSON, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var catalog struct {
		Repositories []string
	}
	err = json.Unmarshal(bodyJSON, &catalog)
	if err != nil {
		return err
	}

	for _, image := range catalog.Repositories {
		tags, err := getImageTags(image)
		if err != nil {
			log.Errorf("Failed to retrieve tags for %s: %s", image, err.Error())
		}
		for _, tag := range tags {
			metadata, err := getImageMetadata(image, tag)
			if err != nil {
				log.Error(err.Error())
			}
			err = installer.Add(image, tag, metadata)
			if err != nil {
				log.Errorf("Could not save installer %s(%s): %s", image, tag, err.Error())
			}
		}
	}

	return nil
}

func processPushEvent(event Event) {
	if event.Target.Tag == "" {
		log.Errorf("Push event for application %s does not containg a tag. Ignoring", event.Target.Repository)
	}
	log.Infof("Processing push event for application %s with tag %s", event.Target.Repository, event.Target.Tag)

	metadata, err := getImageMetadata(event.Target.Repository, event.Target.Tag)
	if err != nil {
		log.Errorf("Could not process image metadata for '%s'(%s): %s", event.Target.Repository, event.Target.Tag, err.Error())
		return
	}

	err = installer.Add(event.Target.Repository, event.Target.Tag, metadata)
	if err != nil {
		log.Errorf("Could not save installer %s(%s): %s", event.Target.Repository, event.Target.Tag, err.Error())
		return
	}
}

// ProcessEvents takes an events array and process all the events of type "push"
func ProcessEvents(events []Event) {
	for _, event := range events {
		if event.Action != "push" {
			log.Debug("Ignoring event of type " + event.Action)
			continue
		}
		log.Info("Received push event from registry")
		go processPushEvent(event)
	}
}
