package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/schollz/progressbar/v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	safewayOauthGetTokenUrl       = "https://albertsons.okta.com/oauth2/ausp6soxrIyPrm8rS2p6/v1/token"
	safewayManufacturerCouponsUrl = "https://nimbus.safeway.com/emmd/service/gallery/offer/mfg?storeId="
	safewayPersonalizedCouponsUrl = "https://nimbus.safeway.com/emmd/service/gallery/offer/pd?storeId="
	safewayAddOfferUrl            = "https://nimbus.safeway.com/Clipping1/services/clip/items?storeId="
	safewayShoppingListUrl        = "https://nimbus.safeway.com/emmd/service/mylist/default/details?storeId="
	safewayClientIdUrl            = "0oap6kkp7Sefg24rB2p6"
	safewayClientSecretUrl        = "4UpmzD4hlF2VYQqYjDUoamgLu2Bo1OzagpfG7yus"
)

type SafewayCouponService struct {
	accessToken string
}

func CreateSafewayCouponService(username, password string) (*SafewayCouponService, error) {
	res := SafewayCouponService{}
	accessToken, err := res.getOauthAccessToken(username, password)
	if err != nil {
		return nil, err
	}
	res.accessToken = accessToken
	return &res, nil
}

type ManufacturerCoupons struct {
	Coupons []Coupon `json:"manufacturerCoupons"`
}

type PersonalizedOffers struct {
	Offers []Offer `json:"personalizedDeals"`
}

type ShoppingList struct {
	Items []ListItem `json:"shoppingList"`
}

type Offer struct {
	OfferID     string `json:"offerID"`
	Description string `json:"description"`
	ItemType    string `json:"offerPgm"`
	Name        string `json:"name"`
}

type Coupon struct {
	CouponID    string `json:"couponID"`
	Description string `json:"description"`
	ItemType    string `json:"offerPgm"`
}

type ListItem struct {
	OfferId  string `json:"offerId"`
	ItemType string `json:"itemType"`
}

var (
	storeId = flag.String("id", "", "your local safeway store id")
	userName = flag.String("u", "", "your safeway.com username/email")
	password = flag.String("p", "", "your safeway.com password")
)

func main() {

	flag.Parse()

	if *storeId == "" || *userName == "" || *password == ""{
		flag.Usage()
		os.Exit(1)
	}
	fmt.Printf("Logging in...\n")
	service, err := CreateSafewayCouponService(*userName,*password)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Getting manufacturer offers...\n")
	mCoupons := ManufacturerCoupons{}
	err = service.GetCouponOffers(safewayManufacturerCouponsUrl+*storeId, &mCoupons)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Getting personal offers...\n")
	poCoupons := PersonalizedOffers{}
	err = service.GetCouponOffers(safewayPersonalizedCouponsUrl+*storeId, &poCoupons)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Getting your existing offers...\n")
	shoppingList := ShoppingList{}
	err = service.GetCouponOffers(safewayShoppingListUrl+*storeId, &shoppingList)
	if err != nil {
		log.Fatal(err)
	}

	addedCoupons := make(map[string]struct{})
	for _,i := range shoppingList.Items{
		addedCoupons[i.OfferId] = struct{}{}
	}

	fmt.Printf("Adding new offers...\n")
	bar := progressbar.New(len(poCoupons.Offers) + len(mCoupons.Coupons))
	for _, c := range mCoupons.Coupons {
		bar.Add(1)
		if _,ok := addedCoupons[c.CouponID];ok{
			continue
		}
		err = service.AddCouponById(*storeId, c.CouponID, c.ItemType)
		if err != nil {
			log.Fatal(err)
		}
	}
	newCoupons := 0
	for _, o := range poCoupons.Offers {
		bar.Add(1)
		if _,ok := addedCoupons[o.OfferID];ok{
			continue
		}
		err = service.AddCouponById(*storeId, o.OfferID, o.ItemType)
		if err != nil {
			log.Fatal(err)
		}
		newCoupons++
	}
	if newCoupons != 0{
		fmt.Printf("\nAdded [%d] offers",newCoupons)
	}else{
		fmt.Printf("\nNo new offers")
	}

	bar.Finish()

}

func (s SafewayCouponService) AddCouponById(storeId string, couponId string, itemType string) error {
	httpClient := http.Client{
		Timeout: time.Duration(5)* time.Second,
	}

	type addCoupon struct {
		ClipType string `json:"clipType"`
		ItemId   string `json:"itemId"`
		ItemType string `json:"itemType"`
	}

	input := map[string][]addCoupon{}
	toAdd := make([]addCoupon, 2)

	toAdd[0] = addCoupon{
		ClipType: "L",
		ItemId:   couponId,
		ItemType: itemType,
	}
	toAdd[1] = addCoupon{
		ClipType: "C",
		ItemId:   couponId,
		ItemType: itemType,
	}

	input["items"] = toAdd
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", safewayAddOfferUrl+storeId, bytes.NewReader(inputBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Safeway/3373 CFNetwork/978.0.7 Darwin/18.6.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to add coupons - code: " + resp.Status)
	}

	return nil
}

func (s SafewayCouponService) GetCouponOffers(url string, object interface{}) error {
	httpClient := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("User-Agent", "Safeway/3373 CFNetwork/978.0.7 Darwin/18.6.0")
	req.AddCookie(&http.Cookie{
		Name:"swyConsumerDirectoryPro",
		Value:s.accessToken,
	})
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK{
		return errors.New("non 200 response "+resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&object)
	if err != nil {
		return err
	}
	return nil
}

func (s SafewayCouponService) getOauthAccessToken(username, password string) (string, error) {
	httpClient := http.Client{}
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("grant_type", "password")
	data.Set("scope", "openid profile offline_access")

	req, err := http.NewRequest("POST", safewayOauthGetTokenUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(safewayClientIdUrl, safewayClientSecretUrl)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Safeway/3373 CFNetwork/978.0.7 Darwin/18.6.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	oauthResponse := map[string]interface{}{}
	err = json.Unmarshal(body, &oauthResponse)
	if err != nil {
		return "", err
	}
	return oauthResponse["access_token"].(string), nil
}
