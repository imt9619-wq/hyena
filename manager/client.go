package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/imt9619-wq/hyena/manager/handler"
	"github.com/sandertv/gophertunnel/minecraft"
	"golang.org/x/oauth2"
)

type ClientTokenIO interface {
	FetchClients() []ClientConfig
	SaveClients(ClientConfig)
}

type ClientConfig struct {
	TokenTag string   // simple tag for this client, will not be modified by the library
	Token *oauth2.Token
}

type Client struct{
	conf *ClientConfig
	id uuid.UUID 
	outgoingConn chan *ClientConn
	closeConnChan chan uuid.UUID // client will send them conn uuid when conn is closing
	managerClosed <-chan struct{}
}


func (c ClientConfig) new(cc chan *ClientConn, ccc chan uuid.UUID, closed <-chan struct{}) *Client{
	clt := &Client{}
	clt.conf = &c
	clt.id = uuid.New()
	clt.outgoingConn = cc
	clt.closeConnChan = ccc
	clt.managerClosed = closed
	return clt
}


type DefaultClientTokenIO struct{
	ClientFolder string
}


func (d *DefaultClientTokenIO) FetchClients() []ClientConfig {
	cConfs := make([]ClientConfig, 0 ,10)

	err := os.MkdirAll(d.ClientFolder, 0755)
	if err != nil {
		log.Fatal(err)
	}

	entries, _ := os.ReadDir(d.ClientFolder)	
	for _, entry := range entries{
		clientJsonPath := filepath.Join(d.ClientFolder, entry.Name())
		data, err := os.ReadFile(clientJsonPath)
		if err != nil || !json.Valid(data){
			continue
		}

		var tok oauth2.Token
		if err := json.Unmarshal(data, &tok); err != nil{
			fmt.Printf("Error when parsing json on %s : %v\n", entry.Name(), err)
			continue
		}
		tagName := strings.TrimSuffix(entry.Name(), ".json")
		cConfs = append(cConfs, ClientConfig{TokenTag: tagName, Token: &tok})
	}

	return cConfs
}


func (d *DefaultClientTokenIO) SaveClients(cConf ClientConfig) {
	err := os.MkdirAll(d.ClientFolder, 0755)
	if err != nil {
		log.Fatal(err)
	}

	if cConf.Token == nil {
		return
	}
	clientJsonPath := filepath.Join(d.ClientFolder, cConf.TokenTag+".json")
	dir := filepath.Dir(clientJsonPath)
	_ = os.MkdirAll(dir, 0o700)
	b, err := json.Marshal(cConf.Token)
	if err != nil {
		return
	}
	_ = os.WriteFile(clientJsonPath, b, 0o600)
	fmt.Printf("%s token saved\n", cConf.TokenTag)
}


func (c *Client) newClientConn(conn *minecraft.Conn, h handler.ConnHandler) *ClientConn{
	return &ClientConn{
		connBuf: handler.NewConnBuf(conn, h),
		client: c,
		id: uuid.New(),
	}
}