package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type VaporClient struct {
	apiToken string
	apiHost  string

	Http http.Client
}

type ErrorResponse struct {
	Message string
}

func prepareRequest[T interface{}](client *VaporClient, method string, path string, decode *T, body io.Reader) error {
	apiHost := client.apiHost

	if apiHost == "" {
		apiHost = defaultApiHost
	}

	baseUrl, err := url.Parse(apiHost)

	if err != nil {
		log.Fatal(err)
	}

	uri := baseUrl.JoinPath(path).String()

	req, reqErr := http.NewRequest(method, uri, body)

	if reqErr != nil {
		return reqErr
	}

	req.Header.Add("Authorization", "Bearer "+client.apiToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, resErr := client.Http.Do(req)

	if resErr != nil {
		return resErr
	}

	if res.StatusCode > 299 {
		errorRes := ErrorResponse{}

		json.NewDecoder(res.Body).Decode(&errorRes)

		return errors.New(strconv.Itoa(res.StatusCode) + " " + method + " request to " + uri + " failed with message: " + errorRes.Message)
	}

	decodeErr := json.NewDecoder(res.Body).Decode(&decode)

	// resBody, _ := io.ReadAll(res.Body)

	// fmt.Print(resBody)

	return decodeErr
}

type Account struct {
	Id              int    `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	Email           string `json:"email,omitempty"`
	EmailVerifiedAt string `json:"email_verified_at,omitempty"`
	AddressLineOne  string `json:"address_line_one,omitempty"`
	Teams           []Team `json:"teams,omitempty"`
	AvatarUrl       string `json:"avatar_url,omitempty"`
	Sandboxed       bool   `json:"is_sandboxed,omitempty"`
}

func (client *VaporClient) GetAccount() (*Account, error) {
	account := Account{}

	err := prepareRequest(client, "GET", "api/user", &account, nil)

	return &account, err
}

type Team struct {
	Id                       int     `json:"id,omitempty"`
	Name                     string  `json:"name,omitempty"`
	AwsId                    string  `json:"aws_external_id,omitempty"`
	SentryOrganisationName   string  `json:"sentry_organization_name,omitempty"`
	SentryOrganisationRegion string  `json:"sentry_organization_region,omitempty"`
	Owner                    Account `json:"owner,omitempty"`
}

func (client *VaporClient) GetTeams() ([]Team, error) {
	teams := []Team{}

	err := prepareRequest(client, "GET", "api/teams", &teams, nil)

	return teams, err
}

func (client *VaporClient) CreateTeam(team Team) (*Team, error) {
	createdTeam := Team{}

	// Fixes the empty owner object sent to API even using omitempty
	val, _ := json.Marshal(struct {
		Name string `json:"name"`
	}{
		Name: team.Name,
	})

	err := prepareRequest(client, "POST", "api/owned-teams", &createdTeam, bytes.NewBuffer(val))

	return &createdTeam, err
}

func (client *VaporClient) GetTeamMembers(teamId int) ([]Account, error) {
	members := []Account{}

	err := prepareRequest(client, "GET", "api/teams/"+strconv.Itoa(teamId)+"/members", &members, nil)

	return members, err
}

func (client *VaporClient) AddTeamMember(teamId int, email string, permissions []string) (*Account, error) {
	createdUser := Account{}

	// Fixes the empty owner object sent to API even using omitempty
	val, _ := json.Marshal(struct {
		Email       string   `json:"email"`
		Permissions []string `json:"permissions"`
	}{
		Email:       email,
		Permissions: permissions,
	})

	err := prepareRequest(client, "POST", "api/teams/"+strconv.Itoa(teamId)+"/members", &createdUser, bytes.NewBuffer(val))

	return &createdUser, err
}

func (client *VaporClient) RemoveTeamMember(teamId int, email string) (*Account, error) {
	createdUser := Account{}

	// Fixes the empty owner object sent to API even using omitempty
	val, _ := json.Marshal(struct {
		Email string `json:"email"`
	}{
		Email: email,
	})

	err := prepareRequest(client, "DELETE", "api/teams/"+strconv.Itoa(teamId)+"/members", &createdUser, bytes.NewBuffer(val))

	return &createdUser, err
}

type VaporProvider struct {
	Id                    int    `json:"id,omitempty"`
	TeamId                int    `json:"team_id,omitempty"`
	Type                  string `json:"type,omitempty"`
	Name                  string `json:"name,omitempty"`
	Uuid                  string `json:"uuid,omitempty"`
	RoleArn               string `json:"role_arn,omitempty"`
	RoleSync              bool   `json:"role_sync,omitempty"`
	SnsTopicArn           string `json:"sns_topic_arn,omitempty"`
	NetworkLimit          int    `json:"network_limit,omitempty"`
	LastDeletedRestApiAt  string `json:"last_deleted_rest_api_at,omitempty"`
	QueuedForDeletion     bool   `json:"queued_for_deletion,omitempty"`
	Concurrency           int    `json:"concurrency,omitempty"`
	UnreservedConcurrency int    `json:"unreserved_concurrency,omitempty"`
}

type VaporProviderMeta struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

func (client *VaporClient) CreateProvider(teamId int, provider VaporProvider, key string, secret string) error {
	val, _ := json.Marshal(struct {
		Type string            `json:"type"`
		Name string            `json:"name"`
		Meta VaporProviderMeta `json:"meta"`
	}{
		Type: provider.Type,
		Name: provider.Name,
		Meta: VaporProviderMeta{
			Key:    key,
			Secret: secret,
		},
	})

	err := prepareRequest(client, "POST", "api/teams/"+strconv.Itoa(teamId)+"/providers", &VaporProvider{}, bytes.NewBuffer(val))

	return err
}

func (client *VaporClient) GetProviders(teamId int) ([]VaporProvider, error) {
	providers := []VaporProvider{}

	err := prepareRequest(client, "GET", "api/teams/"+strconv.Itoa(teamId)+"/providers", &providers, nil)

	return providers, err
}

func (client *VaporClient) RemoveProvider(providerId int) error {
	err := prepareRequest(client, "DELETE", "api/providers/"+strconv.Itoa(providerId), &VaporProvider{}, nil)

	return err
}

type VaporZone struct {
	Id                int           `json:"id,omitempty"`
	TeamId            int           `json:"team_id,omitempty"`
	CloudProviderId   int           `json:"cloud_provider_id,omitempty"`
	ZoneId            string        `json:"zone_id,omitempty"`
	Zone              string        `json:"zone,omitempty"`
	Nameservers       []string      `json:"nameservers,omitempty"`
	SesVerified       bool          `json:"ses_verified,omitempty"`
	Imporing          bool          `json:"importing,omitempty"`
	QueuedForDeletion int           `json:"queued_for_deletion,omitempty"`
	RecordsCount      int           `json:"records_count,omitempty"`
	CloudProvider     VaporProvider `json:"cloud_provider,omitempty"`
}

func (client *VaporClient) GetZones(teamId int) ([]VaporZone, error) {
	zones := []VaporZone{}

	err := prepareRequest(client, "GET", "api/teams/"+strconv.Itoa(teamId)+"/zones", &zones, nil)

	return zones, err
}

func (client *VaporClient) GetZone(zoneId int) (VaporZone, error) {
	zone := VaporZone{}

	err := prepareRequest(client, "GET", "api/zones/"+strconv.Itoa(zoneId), &zone, nil)

	return zone, err
}

func (client *VaporClient) CreateZone(teamId int, providerId int, name string) (VaporZone, error) {
	zone := VaporZone{}

	val, _ := json.Marshal(struct {
		CloudProviderId int    `json:"cloud_provider_id"`
		Zone            string `json:"zone"`
	}{
		CloudProviderId: providerId,
		Zone:            name,
	})

	err := prepareRequest(client, "POST", "api/teams/"+strconv.Itoa(teamId)+"/zones", &zone, bytes.NewBuffer(val))

	return zone, err
}

func (client *VaporClient) RemoveZone(zoneId int) error {
	err := prepareRequest(client, "DELETE", "api/zones/"+strconv.Itoa(zoneId), &VaporZone{}, nil)

	return err
}

type VaporZoneRecord struct {
	Id     int    `json:"id,omitempty"`
	ZoneId int    `json:"zone_id,omitempty"`
	Type   string `json:"type,omitempty"`
	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
}

func (client *VaporClient) CreateZoneRecord(record VaporZoneRecord) (VaporZoneRecord, error) {
	zoneRecord := VaporZoneRecord{}

	val, _ := json.Marshal(record)

	err := prepareRequest(client, "POST", "api/zones/"+strconv.Itoa(record.ZoneId)+"/records", &zoneRecord, bytes.NewBuffer(val))

	return zoneRecord, err
}

func (client *VaporClient) RemoveZoneRecord(record VaporZoneRecord) error {
	err := prepareRequest(client, "DELETE", "api/zones/"+strconv.Itoa(record.ZoneId)+"/records?type="+record.Type+"&name="+record.Name+"&value="+record.Value, &VaporZone{}, nil)

	return err
}
