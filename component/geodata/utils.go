package geodata

import (
	"github.com/Dreamacro/clash/component/geodata/router"
	"strings"
)

var geoLoaderName = "memconservative"

//  geoLoaderName = "standard"

func LoaderName() string {
	return geoLoaderName
}

func SetLoader(newLoader string) {
	geoLoaderName = newLoader
}

func LoadGeoSiteMatcher(countryCode string) (*router.DomainMatcher, int, error) {
	geoLoader, err := GetGeoDataLoader(geoLoaderName)
	if err != nil {
		return nil, 0, err
	}

	domains, err := geoLoader.LoadGeoSite(countryCode)
	if err != nil {
		return nil, 0, err
	}

	/**
	linear: linear algorithm
	matcher, err := router.NewDomainMatcher(domains)
	mph：minimal perfect hash algorithm
	*/
	matcher, err := router.NewMphMatcherGroup(domains)
	if err != nil {
		return nil, 0, err
	}

	return matcher, len(domains), nil
}

func LoadGeoIPMatcher(country string) (*router.GeoIPMatcher, int, error) {
	geoLoader, err := GetGeoDataLoader(geoLoaderName)
	if err != nil {
		return nil, 0, err
	}

	records, err := geoLoader.LoadGeoIP(strings.ReplaceAll(country, "!", ""))
	if err != nil {
		return nil, 0, err
	}

	geoIP := &router.GeoIP{
		CountryCode:  country,
		Cidr:         records,
		ReverseMatch: strings.Contains(country, "!"),
	}

	matcher, err := router.NewGeoIPMatcher(geoIP)
	if err != nil {
		return nil, 0, err
	}

	return matcher, len(records), nil
}