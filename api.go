package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// ScalewayAPI is the interface used to communicate with the Scaleway API
type ScalewayAPI struct {
	// APIEndpoint is the endpoint to the Scaleway API
	APIEndPoint string

	// Organization is the identifier of the Scaleway orgnization
	Organization string

	// Token is the authentication token for the Scaleway organization
	Token string

	// Cache is used to quickly resolve identifiers from names
	Cache *ScalewayCache
}

// ScalewayAPIError represents a Scaleway API Error
type ScalewayAPIError struct {
	// Message is a human-friendly error message
	ApiMessage string `json:"message,omitempty"`

	// Type is a string code that defines the kind of error
	Type string `json:"type,omitempty"`

	// StatusCode is the HTTP status code received
	StatusCode int `json:"-"`

	// Message
	Message string `json:"-"`
}

func (e ScalewayAPIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.ApiMessage != "" {
		return e.ApiMessage
	}
	if e.StatusCode != 0 {
		return fmt.Sprintf("invalid return code, got %d", e.StatusCode)
	}
	panic(e)
}

func (e ScalewayAPIError) Debug() {
	log.WithFields(log.Fields{
		"StatusCode": e.StatusCode,
		"Type":       e.Type,
		"Message":    e.Message,
	}).Debug(e.ApiMessage)
}

// ScalewayIPAddress represents a Scaleway IP address
type ScalewayIPAddress struct {
	// IP is an IPv4 address
	IP string `json:"address,omitempty"`
}

// ScalewayVolume represents a Scaleway Volume
type ScalewayVolume struct {
	// Identifier is a unique identifier for the volume
	Identifier string `json:"id,omitempty"`

	// Size is allocated size of the volume
	Size int64 `json:"size,omitempty"`

	// CreationDate is the creation date of the volume
	CreationDate string `json:"creation_date,omitempty"`

	// ModificationDate is the date of the last modification of the volume
	ModificationDate string `json:"modification_date,omitempty"`

	// Name is the name of the volume
	Name string `json:"name,omitempty"`
}

// ScalewayImage represents a Scaleway Image
type ScalewayImage struct {
	// Identifier is a unique identifier for the image
	Identifier string `json:"id,omitempty"`

	// Name is a user-defined name for the image
	Name string `json:"name,omitempty"`

	// CreationDate is the creation date of the image
	CreationDate string `json:"creation_date,omitempty"`

	// ModificationDate is the date of the last modification of the image
	ModificationDate string `json:"modification_date,omitempty"`

	// RootVolume is the root volume bound to the image
	RootVolume ScalewayVolume `json:"root_volume,omitempty"`
}

// ScalewayOneImage represents the response of a GET /images/UUID API call
type ScalewayOneImage struct {
	Image ScalewayImage `json:"image,omitempty"`
}

// ScalewayImages represents a group of Scaleway images
type ScalewayImages struct {
	// Images holds scaleway images of the response
	Images []ScalewayImage `json:"images,omitempty"`
}

// ScalewaySnapshot represents a Scaleway Snapshot
type ScalewaySnapshot struct {
	// Identifier is a unique identifier for the snapshot
	Identifier string `json:"id,omitempty"`

	// Name is a user-defined name for the snapshot
	Name string `json:"name,omitempty"`

	// CreationDate is the creation date of the snapshot
	CreationDate string `json:"creation_date,omitempty"`

	// ModificationDate is the date of the last modification of the snapshot
	ModificationDate string `json:"modification_date,omitempty"`

	// Size is allocated size of the volume
	Size int64 `json:"size,omitempty"`

	// Organization is the owner of the snapshot
	Organization string `json:"organization"`

	// State is the current state of the snapshot
	State string `json:"state"`

	// VolumeType is the kind of volume behind the snapshot
	VolumeType string `json:"volume_type"`

	// BaseVolume is the volume from which the snapshot inherits
	BaseVolume ScalewayVolume `json:"base_volume,omitempty"`
}

// ScalewayOneSnapshot represents the response of a GET /images/UUID API call
type ScalewayOneSnapshot struct {
	Snapshot ScalewaySnapshot `json:"snapshot,omitempty"`
}

// ScalewaySnapshots represents a group of Scaleway snapshots
type ScalewaySnapshots struct {
	// Snapshots holds scaleway snapshots of the response
	Snapshots []ScalewaySnapshot `json:"snapshots,omitempty"`
}

// ScalewayBootCmdArgs represents the boot arguments of a bootscript
type ScalewayBootCmdArgs struct {
	// Identifier is the unique identifier of boot args
	Identifier string `json:"id,omitempty"`

	// Value is the content of the cmd args
	Value string `json:"value,omitempty"`
}

// ScalewayInitrd represents the initrd used by a bootscript
type ScalewayInitrd struct {
	// Identifier is the unique identifier of the initrd
	Identifier string `json:"id,omitempty"`

	// Path is the path to the initrd used
	Path string `json:"path,omitempty"`

	// Title is the title of the initrd used
	Title string `json:"title,omitempty"`
}

// ScalewayKernel represents a kernel used on C1 servers
type ScalewayKernel struct {
	// Identifier is the unique identifier of the kernel
	Identifier string `json:"id,omitempty"`

	// DTB is the kernel DTB used by this kernel
	DTB string `json:"dtb,omitempty"`

	// Path is the path to the kernel image
	Path string `json:"path,omitempty"`

	// Title is the title of the kernel
	Title string `json:"title,omitempty"`
}

// ScalewayBootscript represents a Scaleway Bootscript
type ScalewayBootscript struct {
	// Identifier is a unique identifier for the bootscript
	Identifier string `json:"id,omitempty"`

	// Name is a user-defined name for the bootscript
	Title string `json:"title,omitempty"`

	// BootCmdArgs represents the arguments used to boot
	BootCmdArgs ScalewayBootCmdArgs `json:"bootcmdargs,omitempty"`

	// Initrd is the initrd used by this bootscript
	Initrd ScalewayInitrd `json:initrd,omitempty`

	// Kernel is the kernel associated to this server
	Kernel ScalewayKernel `json:"kernel,omitempty"`
}

// ScalewayOneBootscript represents the response of a GET /images/UUID API call
type ScalewayOneBootscript struct {
	Bootscript ScalewayBootscript `json:"bootscript,omitempty"`
}

// ScalewayBootscripts represents a group of Scaleway bootscripts
type ScalewayBootscripts struct {
	// Bootscripts holds scaleway bootscripts of the response
	Bootscripts []ScalewayBootscript `json:"bootscripts,omitempty"`
}

// ScalewayServer represents a Scaleway C1 server
type ScalewayServer struct {
	// Identifier is a unique identifier for the server
	Identifier string `json:"id,omitempty"`

	// Name is the user-defined name of the server
	Name string `json:"name,omitempty"`

	// CreationDate is the creation date of the server
	CreationDate string `json:"creation_date,omitempty"`

	// ModificationDate is the date of the last modification of the server
	ModificationDate string `json:"modification_date,omitempty"`

	// Image is the image used by the server
	Image ScalewayImage `json:"image,omitempty"`

	// PublicIP is the public IP address bound to the server
	PublicAddress ScalewayIPAddress `json:"public_ip,omitempty"`

	// State is the current status of the server
	State string `json:"state,omitempty"`
}

// ScalewayServer represents a Scaleway C1 server definition
type ScalewayServerDefinition struct {
	// Name is the user-defined name of the server
	Name string `json:"name"`

	// Image is the image used by the server
	Image string `json:"image"`

	// Bootscript is the bootscript used by the server
	Bootscript *string `json:"bootscript"`

	// Tags are the metadata tags attached to the server
	// Tags []string `json:"tags",omitempty`

	// Organization is the owner of the server
	Organization string `json:"organization"`
}

// ScalewayOneServer represents the response of a GET /servers/UUID API call
type ScalewayOneServer struct {
	Server ScalewayServer `json:"server,omitempty"`
}

// ScalewayServers represents a group of Scaleway C1 servers
type ScalewayServers struct {
	// Servers holds scaleway servers of the response
	Servers []ScalewayServer `json:"servers,omitempty"`
}

// ScalewayServerAction represents an action to perform on a Scaleway C1 server
type ScalewayServerAction struct {
	// State is the current status of the server
	Action string `json:"action,omitempty"`
}

// NewScalewayAPI creates a ready-to-use ScalewayAPI client
func NewScalewayAPI(endpoint, organization, token string) (*ScalewayAPI, error) {
	cache, err := NewScalewayCache()
	if err != nil {
		return nil, err
	}
	return &ScalewayAPI{
		APIEndPoint:  endpoint,
		Organization: organization,
		Token:        token,
		Cache:        cache,
	}, nil
}

// Sync flushes out the cache to the disk
func (s *ScalewayAPI) Sync() {
	s.Cache.Save()
}

// GetResponse returns a http.Response object for the requested resource
func (s *ScalewayAPI) GetResponse(resource string) (*http.Response, error) {
	uri := fmt.Sprintf("%s/%s", strings.TrimRight(s.APIEndPoint, "/"), resource)
	log.Debugf("GET %s", uri)
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", s.Token)
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

// PostResponse returns a http.Response object for the updated resource
func (s *ScalewayAPI) PostResponse(resource string, data interface{}) (*http.Response, error) {
	uri := fmt.Sprintf("%s/%s", strings.TrimRight(s.APIEndPoint, "/"), resource)
	client := &http.Client{}
	payload := new(bytes.Buffer)
	encoder := json.NewEncoder(payload)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	log.Debugf("POST %s payload=%s", uri, payload)
	req, err := http.NewRequest("POST", uri, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", s.Token)
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

// GetServers get the list of servers from the ScalewayAPI
func (s *ScalewayAPI) GetServers(all bool, limit int) (*[]ScalewayServer, error) {
	query := url.Values{}
	if !all {
		query.Set("state", "running")
	}
	if limit > 0 {
		// FIXME: wait for the API to be ready
		// query.Set("per_page", strconv.Itoa(limit))
	}
	resp, err := s.GetResponse("servers?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var servers ScalewayServers
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&servers)
	if err != nil {
		return nil, err
	}
	for _, server := range servers.Servers {
		s.Cache.InsertServer(server.Identifier, server.Name)
	}
	// FIXME: when api limit is ready, remove the following code
	if limit > 0 && limit < len(servers.Servers) {
		servers.Servers = servers.Servers[0:limit]
	}
	return &servers.Servers, nil
}

// GetServer get a server from the ScalewayAPI
func (s *ScalewayAPI) GetServer(serverId string) (*ScalewayServer, error) {
	resp, err := s.GetResponse("servers/" + serverId)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var oneServer ScalewayOneServer
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&oneServer)
	if err != nil {
		return nil, err
	}
	s.Cache.InsertServer(oneServer.Server.Identifier, oneServer.Server.Name)
	return &oneServer.Server, nil
}

// PostServerAction posts an action on a server
func (s *ScalewayAPI) PostServerAction(server_id, action string) error {
	data := ScalewayServerAction{
		Action: action,
	}
	resp, err := s.PostResponse(fmt.Sprintf("servers/%s/action", server_id), data)
	if err != nil {
		return err
	}

	// Succeed POST code
	if resp.StatusCode == 202 {
		return nil
	}

	var error ScalewayAPIError
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&error)
	if err != nil {
		return err
	}

	error.StatusCode = resp.StatusCode
	error.Debug()
	return error
}

// PostServer create a new server
func (s *ScalewayAPI) PostServer(definition ScalewayServerDefinition) (string, error) {
	definition.Organization = s.Organization
	resp, err := s.PostResponse(fmt.Sprintf("servers"), definition)
	if err != nil {
		return "", err
	}

	// Succeed POST code
	if resp.StatusCode == 201 {
		var server ScalewayOneServer
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&server)
		if err != nil {
			return "", err
		}
		return server.Server.Identifier, nil
	}

	var error ScalewayAPIError
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&error)

	if err != nil {
		return "", err
	}

	error.StatusCode = resp.StatusCode
	error.Debug()
	return "", error
}

// ResolveServer attempts the find a matching Identifier for the input string
func (s *ScalewayAPI) ResolveServer(needle string) ([]string, error) {
	servers := s.Cache.LookUpServers(needle)
	if len(servers) == 0 {
		_, err := s.GetServers(true, 0)
		if err != nil {
			return nil, err
		}
		servers = s.Cache.LookUpServers(needle)
	}
	return servers, nil
}

// ResolveImage attempts the find a matching Identifier for the input string
func (s *ScalewayAPI) ResolveImage(needle string) ([]string, error) {
	images := s.Cache.LookUpImages(needle)
	if len(images) == 0 {
		_, err := s.GetImages()
		if err != nil {
			return nil, err
		}
		images = s.Cache.LookUpImages(needle)
	}
	return images, nil
}

// ResolveBootscript attempts the find a matching Identifier for the input string
func (s *ScalewayAPI) ResolveBootscript(needle string) ([]string, error) {
	bootscripts := s.Cache.LookUpBootscripts(needle)
	if len(bootscripts) == 0 {
		_, err := s.GetBootscripts()
		if err != nil {
			return nil, err
		}
		bootscripts = s.Cache.LookUpBootscripts(needle)
	}
	return bootscripts, nil
}

// GetImages get the list of images from the ScalewayAPI
func (s *ScalewayAPI) GetImages() (*[]ScalewayImage, error) {
	query := url.Values{}
	resp, err := s.GetResponse("images?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var images ScalewayImages
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&images)
	if err != nil {
		return nil, err
	}
	for _, image := range images.Images {
		s.Cache.InsertImage(image.Identifier, image.Name)
	}
	return &images.Images, nil
}

// GetImage gets an image from the ScalewayAPI
func (s *ScalewayAPI) GetImage(imageId string) (*ScalewayImage, error) {
	resp, err := s.GetResponse("images/" + imageId)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var oneImage ScalewayOneImage
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&oneImage)
	if err != nil {
		return nil, err
	}
	s.Cache.InsertImage(oneImage.Image.Identifier, oneImage.Image.Name)
	return &oneImage.Image, nil
}

// GetSnapshots get the list of snapshots from the ScalewayAPI
func (s *ScalewayAPI) GetSnapshots() (*[]ScalewaySnapshot, error) {
	query := url.Values{}
	resp, err := s.GetResponse("snapshots?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var snapshots ScalewaySnapshots
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&snapshots)
	if err != nil {
		return nil, err
	}
	for _, snapshot := range snapshots.Snapshots {
		s.Cache.InsertSnapshot(snapshot.Identifier, snapshot.Name)
	}
	return &snapshots.Snapshots, nil
}

// GetSnapshot gets a snapshot from the ScalewayAPI
func (s *ScalewayAPI) GetSnapshot(snapshotId string) (*ScalewaySnapshot, error) {
	resp, err := s.GetResponse("snapshots/" + snapshotId)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var oneSnapshot ScalewayOneSnapshot
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&oneSnapshot)
	if err != nil {
		return nil, err
	}
	s.Cache.InsertSnapshot(oneSnapshot.Snapshot.Identifier, oneSnapshot.Snapshot.Name)
	return &oneSnapshot.Snapshot, nil
}

// GetBootscripts get the list of bootscripts from the ScalewayAPI
func (s *ScalewayAPI) GetBootscripts() (*[]ScalewayBootscript, error) {
	query := url.Values{}
	resp, err := s.GetResponse("bootscripts?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var bootscripts ScalewayBootscripts
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&bootscripts)
	if err != nil {
		return nil, err
	}
	for _, bootscript := range bootscripts.Bootscripts {
		s.Cache.InsertBootscript(bootscript.Identifier, bootscript.Title)
	}
	return &bootscripts.Bootscripts, nil
}

// GetBootscript gets a bootscript from the ScalewayAPI
func (s *ScalewayAPI) GetBootscript(bootscriptId string) (*ScalewayBootscript, error) {
	resp, err := s.GetResponse("bootscripts/" + bootscriptId)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var oneBootscript ScalewayOneBootscript
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&oneBootscript)
	if err != nil {
		return nil, err
	}
	s.Cache.InsertBootscript(oneBootscript.Bootscript.Identifier, oneBootscript.Bootscript.Title)
	return &oneBootscript.Bootscript, nil
}