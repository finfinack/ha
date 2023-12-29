package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/finfinack/ha/configuration"
	"github.com/finfinack/ha/data"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	ttlcache "github.com/jellydator/ttlcache/v2"
)

var (
	port    = flag.Int("port", 8080, "Listening port for webserver.")
	tlsCert = flag.String("tlsCert", "", "Path to TLS Certificate. If this and -tlsKey is specified, service runs as TLS server.")
	tlsKey  = flag.String("tlsKey", "", "Path to TLS Key. If this and -tlsCert is specified, service runs as TLS server.")

	cacheTTL       = flag.Duration("cacheTTL", 3*time.Hour, "Duration for which to keep the entries in cache.")
	haStatusReload = flag.Duration("haStatusReload", time.Minute, "Duration after which to fetch HA status again.")

	configFile = flag.String("config", "", "Absolute path of the configuration file to use.")
)

const (
	collectEndpoint = "/measure/v1/collect"

	deviceClassTemperature = "temperature"
)

type HAServer struct {
	Cache  *ttlcache.Cache
	Server *http.Server
}

func (ha *HAServer) collectHandler(ctx *gin.Context) {
	entities := []*data.HAEntity{}
	for _, e := range ha.Cache.GetItems() {
		entities = append(entities, e.(*data.HAEntity))
	}

	out, err := convertEntities(entities)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, out)
}

func updateHAStatus(config *configuration.Config, cache *ttlcache.Cache) error {
	req, err := http.NewRequest(http.MethodGet, config.HAStatusURL, nil)
	if err != nil {
		return fmt.Errorf("unable to create request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.HAAuthToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to make request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got unexpected HTTP status: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %s", err)
	}

	var haEntities []*data.HAEntity
	if err := json.Unmarshal(respBody, &haEntities); err != nil {
		return fmt.Errorf("unable to parse JSON response: %s", err)
	}

	filteredEntities, err := filterEntities(haEntities, config.IncludeEntities, config.ExcludeEntities)
	if err != nil {
		return fmt.Errorf("unable to filter HA entities: %s", err)
	}

	for _, e := range filteredEntities {
		cache.Set(e.ID, e)
	}

	return nil
}

func filterEntities(in []*data.HAEntity, includes []string, excludes []string) ([]*data.HAEntity, error) {
	out := []*data.HAEntity{}
	for _, e := range in {
		var err error
		incl, excl := false, false
		for _, f := range includes {
			incl, err = regexp.MatchString(f, e.ID)
			if err != nil {
				return nil, err
			}
			if incl {
				break
			}
		}
		for _, f := range excludes {
			excl, err = regexp.MatchString(f, e.ID)
			if err != nil {
				return nil, err
			}
			if excl {
				break
			}
		}
		if incl && !excl {
			out = append(out, e)
		}
	}
	return out, nil
}

func convertEntities(in []*data.HAEntity) (*data.M5Data, error) {
	out := &data.M5Data{
		LastUpdated: time.Now().Unix(),
	}

	for _, e := range in {
		if strings.ToLower(e.Attributes.DeviceClass) != deviceClassTemperature {
			continue
		}
		temp, err := strconv.ParseFloat(e.State, 32)
		if err != nil {
			continue
		}
		name := e.Attributes.FriendlyName
		name = strings.TrimSuffix(name, " Temperature")
		name = strings.TrimSuffix(name, " temperature")
		room := &data.M5Room{
			Name:        name,
			Temperature: float32(temp),
		}
		out.Rooms = append(out.Rooms, room)
	}

	sort.SliceStable(out.Rooms, func(i, j int) bool {
		return out.Rooms[i].Name < out.Rooms[j].Name
	})

	return out, nil
}

func main() {
	// Set defaults for glog flags. Can be overridden via cmdline.
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "WARNING")
	flag.Set("v", "1")
	// Parse flags globally.
	flag.Parse()

	if *configFile == "" {
		glog.Exit("-config needs to be set")
	}
	config, err := configuration.Read(*configFile)
	if err != nil {
		glog.Exitf("unable to read config file from %q: %s", *configFile, err)
	}

	cache := ttlcache.NewCache()
	cache.SetTTL(time.Duration(*cacheTTL))

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.SetFuncMap(template.FuncMap{})

	srv := HAServer{
		Cache: cache,
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", *port),
			Handler: router, // use `http.DefaultServeMux`
		},
	}
	router.GET(collectEndpoint, srv.collectHandler)

	go func() {
		for {
			if err := updateHAStatus(config, cache); err != nil {
				glog.Warningf("unable to fetch HA status: %s", err)
			}
			time.Sleep(*haStatusReload)
		}
	}()

	if *tlsCert != "" && *tlsKey != "" {
		router.RunTLS(fmt.Sprintf(":%d", *port), *tlsCert, *tlsKey)
	} else {
		router.Run(fmt.Sprintf(":%d", *port))
	}
}
