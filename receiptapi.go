package main

import (
    //"fmt"
    "log"
    "net/http"
    "encoding/json"
    "math"
    "strconv"
    "strings"
    "github.com/google/uuid"
)

//Thank you, https://medium.com/@saharat.paynok/how-to-check-if-the-character-is-alphanumeric-in-go-6783b92ec412 !
func isAlphaNumeric(c byte) bool {
	// Check if the byte value falls within the range of alphanumeric characters
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
type Receipt struct {
	Retailer string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Total string `json:"total"`
	Points int
	Items        []struct {
		ShortDescription string `json:"shortDescription"`
		Price            string `json:"price"`
	} `json:"items"`
}

func main() {
	var receiptsAndPoints map[string]int
	receiptsAndPoints=make(map[string]int)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("<h1>Do you have a receipt?</h1>"))
    })
    //This should put it in a database fit for a json with its own ID, it should create a field for how many points it has
    http.HandleFunc("POST /receipts/process", func(w http.ResponseWriter, r *http.Request) {
        decoder := json.NewDecoder(r.Body)
        decoder.DisallowUnknownFields()
        var rec Receipt
        err:=decoder.Decode(&rec)
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Printf("The retailer is %s and the purchase date is %s\n",rec.Retailer, rec.PurchaseDate)
  
		//* 10 points if the time of purchase is after 2:00pm and before 4:00pm.
		timeParts:=strings.Split(rec.PurchaseTime, ":")
		if timeParts[0]=="14" && timeParts[1]!="00" || timeParts[0]=="15"{
			//fmt.Printf("10 points if the time of purchase is after 2:00pm and before 4:00pm.\n")
			rec.Points+=10
		}
		
		//* 6 points if the day in the purchase date is odd.
		dateParts:=strings.Split(rec.PurchaseDate, "-")
		dayDatePart, err := strconv.Atoi(dateParts[2])
		if err!=nil{
			// do something sensible, return 400 and call the receipt invalid
			w.WriteHeader(400)
			responseMap := map[string]string{"description": "The receipt is invalid"}
			json.NewEncoder(w).Encode(responseMap)
			return
		}
		if dayDatePart%2!=0{
			//fmt.Printf("6 points if the day in the purchase date is odd.\n")
			rec.Points+=6
		}

		//* If the trimmed length of the item description is a multiple of 3, multiply the price by `0.2` and round up to the nearest integer. The result is the number of points earned.
		for i := 0; i < len(rec.Items) ; i++ {
			rec.Items[i].ShortDescription=strings.TrimSpace(rec.Items[i].ShortDescription)
			if len(strings.TrimSpace(rec.Items[i].ShortDescription))%3==0{
				//* 25 points if the total is a multiple of `0.25`.
				itemPriceValue, err := strconv.ParseFloat(rec.Items[i].Price, 64)
				if err != nil {
					// do something sensible, return 400 and call the receipt invalid
					w.WriteHeader(400)
					responseMap := map[string]string{"description": "The receipt is invalid"}
					json.NewEncoder(w).Encode(responseMap)
					return
				}
				//fmt.Printf("If the trimmed length of the item description is a multiple of 3, multiply the price by `0.2` and round up to the nearest integer. The result is the number of points earned.\n")
				itemPrice := float64(itemPriceValue)
				rec.Points+=int(math.Ceil(itemPrice*.2))
			}
		}

		//* 25 points if the total is a multiple of `0.25`.
		value, err := strconv.ParseFloat(rec.Total, 64)
		if err != nil {
			// do something sensible, return 400 and call the receipt invalid
			w.WriteHeader(400)
			responseMap := map[string]string{"description": "The receipt is invalid"}
			json.NewEncoder(w).Encode(responseMap)
			return
		}
		totalFloat := float64(value)
		if math.Mod(totalFloat, .25)==0{
			//fmt.Printf("25 points if the total is a multiple of `0.25`.\n")
			rec.Points+=25
		}
		//* 50 points if the total is a round dollar amount with no cents.
		if math.Mod(totalFloat, 1)==0{
			//fmt.Printf("50 points if the total is a round dollar amount with no cents.\n")
			rec.Points+=50
		}
		//* One point for every alphanumeric character in the retailer name.
		for i := 0; i < len(rec.Retailer) ; i++ {
			if (isAlphaNumeric(rec.Retailer[i])){
				//fmt.Printf("One point for every alphanumeric character in the retailer name.\n")
				rec.Points+=1
			}
		}
		//* 5 points for every two items on the receipt.
		//wait, divide by two strictly, multiply that by 5, then add that to the points?
		//So divide by two, round down, *5, add to points?
		//11/2=5.5, round down to 5, *5 to get 25,
		//rec.Points+=math.Trunc(len(rec.Items)/2.0)*5
		//This feels hacky, how would I do it in Python? Something less respectful of datatypes
		if len(rec.Items)%2==0{
			//fmt.Printf("5 points for every two items on the receipt, which works out to %s.\n", strconv.Itoa(len(rec.Items)/2*5))
			rec.Points+=len(rec.Items)/2*5
		} else{
			//fmt.Printf("5 points for every two items on the receipt, which works out to %s.\n", strconv.Itoa((len(rec.Items)-1)/2*5))
			rec.Points+=(len(rec.Items)-1)/2*5
		}
		//fmt.Printf("That was worth %s points!\n", strconv.Itoa(rec.Points))
		//w.Write([]byte("<p>"+strconv.Itoa(rec.Points)+"</p>\n"))
		uuid := uuid.NewString()
		receiptsAndPoints[uuid]=rec.Points
		responseMap := map[string]string{"id": uuid}
		err= json.NewEncoder(w).Encode(responseMap)

    })  
    http.HandleFunc("GET /receipts/{id}/points", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id:=r.PathValue("id")
		howManyPoints, exists:=receiptsAndPoints[id]
		if !exists{
			// do something sensible, return 400 and call the receipt invalid
			w.WriteHeader(400)
			responseMap := map[string]string{"description": "No receipt found for that id"}
			json.NewEncoder(w).Encode(responseMap)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		responseMap := map[string]int{"points": howManyPoints}
		json.NewEncoder(w).Encode(responseMap)
    })
    log.Fatal(http.ListenAndServe(":8080", nil))
}
