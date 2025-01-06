package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v80"
	"github.com/stripe/stripe-go/v80/paymentintent"
	"github.com/stripe/stripe-go/v80/setupintent"
	"github.com/stripe/stripe-go/v80/webhook"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv.Load: %v", err)
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	http.Handle("/", http.FileServer(http.Dir(os.Getenv("STATIC_DIR"))))
	http.HandleFunc("/create-payment-intent", handleCreatePaymentIntent)
	http.HandleFunc("/resolve-last-payment-intent", handleResolveLastPaymentIntent)
	http.HandleFunc("/create-setup-intent", handleCreateSetupIntent)
	http.HandleFunc("/capture-payment-intent", handleCapturePaymentIntent)
	http.HandleFunc("/cancel-payment-intent", handleCancelPaymentIntent)
	http.HandleFunc("/confirm-payment-intent", handleConfirmPaymentIntent)
	http.HandleFunc("/webhook", handleWebhook)
	http.HandleFunc("/config", handleConfig)

	addr := "localhost:4242"
	log.Printf("Listening on %s ...", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	type Config struct {
		PublishableKey string `json:"publishableKey"`
	}

	cfg := Config{
		PublishableKey: os.Getenv("STRIPE_PUBLISHABLE_KEY"),
	}

	writeJSON(w, cfg)
}

// PayItemParams represents a single item passed from the client.
// In practice, the ID of the PayItemParams object would be some
// ID or reference to an internal product that you can use to
// determine the price. You need to implement calculateOrderAmount
// or a similar function to actually calculate the amount here
// on the server. That way, the user cannot modify the amount that
// is charged by changing the client.
type PayItemParams struct {
	ID string `json:"id"`
}

// PayRequestParams represents the structure of the request from
// the client.
type PayRequestParams struct {
	Currency string          `json:"currency"`
	Items    []PayItemParams `json:"items"`
}

func handleCreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// Decode the incoming request
	req := PayRequestParams{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if !errors.Is(err, io.EOF) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("json.NewDecoder.Decode: %v", err)
			return
		}
		req.Currency = "USD"
	}

	//customerParams := &stripe.CustomerParams{}
	//c, err := customer.New(customerParams)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	log.Printf("customer.New: %v", err)
	//	return
	//}

	// authorize 1 USD to return it back after confirmation - https://docs.stripe.com/payments/place-a-hold-on-a-payment-method#authorize-only
	paymentIntentParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(100),
		Currency: stripe.String(req.Currency),
		Customer: stripe.String("cus_R88nCQ6UTjjC2u"),
		//Customer:                  stripe.String(c.ID),
		CaptureMethod:             stripe.String(string(stripe.PaymentIntentCaptureMethodManual)),
		SetupFutureUsage:          stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession)),
		StatementDescriptor:       stripe.String("firebolt"),
		StatementDescriptorSuffix: stripe.String("pre-auth"),
		Description:               stripe.String("Pre-authorize 1.00 USD to return it back after confirmation"),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(paymentIntentParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("paymentintent.New: %v", err)
		return
	}

	writeJSON(w, struct {
		PublicKey    string `json:"publicKey"`
		ClientSecret string `json:"clientSecret"`
		ID           string `json:"id"`
	}{
		PublicKey:    os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		ClientSecret: pi.ClientSecret,
		ID:           pi.ID,
	})
}
func handleResolveLastPaymentIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// PayRequestParams represents the structure of the request from
	// the client.
	type ResolvePayRequestParams struct {
		CustomerID string `json:"customerID"`
	}
	// Decode the incoming request
	req := ResolvePayRequestParams{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	listPaymentIntent := paymentintent.List(&stripe.PaymentIntentListParams{
		Customer: stripe.String(req.CustomerID),
	})
	if err := listPaymentIntent.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("paymentintent.List: %v", err)
		return
	}
	//for listPaymentIntent.Next() {
	//}

	//// authorize 1 USD to return it back after confirmation - https://docs.stripe.com/payments/place-a-hold-on-a-payment-method#authorize-only
	//paymentIntentParams := &stripe.PaymentIntentParams{
	//	Amount:   stripe.Int64(100),
	//	Currency: stripe.String(req.Currency),
	//	//Customer: stripe.String("cus_R88nCQ6UTjjC2u"),
	//	Customer:                  stripe.String(c.ID),
	//	CaptureMethod:             stripe.String(string(stripe.PaymentIntentCaptureMethodManual)),
	//	SetupFutureUsage:          stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession)),
	//	StatementDescriptor:       stripe.String("firebolt"),
	//	StatementDescriptorSuffix: stripe.String("pre-auth"),
	//	Description:               stripe.String("Pre-authorize 1.00 USD to return it back after confirmation"),
	//	AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
	//		Enabled: stripe.Bool(true),
	//	},
	//}
	//
	//pi, err := paymentintent.New(paymentIntentParams)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	log.Printf("paymentintent.New: %v", err)
	//	return
	//}
	if !listPaymentIntent.Next() {
		http.Error(w, "No payment intent found", http.StatusInternalServerError)
		log.Printf("paymentintent.List: No payment intent found")
		return
	}
	pi := listPaymentIntent.PaymentIntent()
	log.Printf("pi: %v\n", pi)
	if pi.Status == stripe.PaymentIntentStatusRequiresPaymentMethod {
		// should be in webhook handler
		oldPiID := pi.ID
		var err error
		// new needed, so customer can select
		pi, err = paymentintent.New(copyIntentForFreshPayment(pi))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("paymentintent.Update: %v", err)
			return
		}
		_, err = paymentintent.Cancel(oldPiID, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("paymentintent.Cancel: %v", err)
		}
	}
	writeJSON(w, struct {
		Amount       int64  `json:"amount"`
		PublicKey    string `json:"publicKey"`
		ClientSecret string `json:"clientSecret"`
		ID           string `json:"id"`
	}{
		Amount:       pi.Amount,
		PublicKey:    os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		ClientSecret: pi.ClientSecret,
		ID:           pi.ID,
	})
}

func copyIntentForFreshPayment(pi *stripe.PaymentIntent) *stripe.PaymentIntentParams {

	receiptEmail := pi.ReceiptEmail
	var receiptEmailRef *string
	if receiptEmail != "" {
		receiptEmailRef = &receiptEmail
	}
	return &stripe.PaymentIntentParams{
		Params:               stripe.Params{},
		Amount:               stripe.Int64(pi.Amount),
		ApplicationFeeAmount: stripe.Int64(pi.ApplicationFeeAmount),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			AllowRedirects: stripe.String(string(stripe.PaymentIntentAutomaticPaymentMethodsAllowRedirectsNever)),
			Enabled:        stripe.Bool(true),
		},
		CaptureMethod:              stripe.String(string(stripe.PaymentIntentCaptureMethodAutomatic)),
		ClientSecret:               nil,
		Confirm:                    nil,
		ConfirmationMethod:         nil,
		ConfirmationToken:          nil,
		Currency:                   stripe.String(string(pi.Currency)),
		Customer:                   stripe.String(pi.Customer.ID),
		Description:                stripe.String(pi.Description),
		Expand:                     nil,
		Mandate:                    nil,
		MandateData:                nil,
		Metadata:                   pi.Metadata,
		OnBehalfOf:                 nil,
		PaymentMethod:              nil,
		PaymentMethodConfiguration: nil,
		PaymentMethodData:          nil,
		PaymentMethodOptions:       nil,
		PaymentMethodTypes:         nil,
		RadarOptions:               nil,
		ReceiptEmail:               receiptEmailRef,
		ReturnURL:                  nil,
		SetupFutureUsage:           stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession)),
		Shipping:                   nil,
		StatementDescriptor:        stripe.String(pi.StatementDescriptor),
		StatementDescriptorSuffix:  stripe.String(pi.StatementDescriptorSuffix),
		TransferData:               nil,
		TransferGroup:              nil,
		ErrorOnRequiresAction:      stripe.Bool(false),
		OffSession:                 nil,
		UseStripeSDK:               nil,
	}
}

// https://docs.stripe.com/financial-connections/ach-direct-debit-payments#getting-started
// read about ACH Direct Debit Payments optimizations to check accounts balances before charge.
// How to check permissions given in default flow?
func handleCreateSetupIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// Decode the incoming request
	req := PayRequestParams{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	//customerParams := &stripe.CustomerParams{}
	//c, err := customer.New(customerParams)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	log.Printf("customer.New: %v", err)
	//	return
	//}

	setupIntentParams := &stripe.SetupIntentParams{
		Customer: stripe.String("cus_R88nCQ6UTjjC2u"),
		//Customer:    stripe.String(c.ID),
		Description: stripe.String("Capture payment details for future use"),
		AutomaticPaymentMethods: &stripe.SetupIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := setupintent.New(setupIntentParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("paymentintent.New: %v", err)
		return
	}

	writeJSON(w, struct {
		PublicKey    string `json:"publicKey"`
		ClientSecret string `json:"clientSecret"`
		ID           string `json:"id"`
	}{
		PublicKey:    os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		ClientSecret: pi.ClientSecret,
		ID:           pi.ID,
	})
}

// handleCapturePaymentIntent captures amount specified in payment intent by specified id.
// https://docs.stripe.com/payments/place-a-hold-on-a-payment-method#capture-funds
func handleCapturePaymentIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	// CaptureRequestParams represents the structure of the request from
	// the client.
	type CaptureRequestParams struct {
		PaymentIntentID string `json:"paymentIntentID"`
		Amount          int64  `json:"amount"`
	}

	// Decode the incoming request
	req := CaptureRequestParams{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	params := &stripe.PaymentIntentCaptureParams{
		AmountToCapture: stripe.Int64(req.Amount),
	}

	pi, err := paymentintent.Capture(req.PaymentIntentID, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("paymentintent.Capture: %v", err)
	}

	writeJSON(w, pi)
}

// handleConfirmPaymentIntent captures amount specified in payment intent by specified id.
// https://docs.stripe.com/payments/payment-intents/upgrade-to-handle-actions
func handleConfirmPaymentIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	// CaptureRequestParams represents the structure of the request from
	// the client.
	type ConfirmRequestParams struct {
		PaymentIntentID string `json:"paymentIntentID"`
	}

	// Decode the incoming request
	req := ConfirmRequestParams{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	pi, err := paymentintent.Get(req.PaymentIntentID, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("paymentintent.Get: %v", err)
		return
	}

	writeJSON(w, struct {
		PublicKey    string `json:"publicKey"`
		ClientSecret string `json:"clientSecret"`
		ID           string `json:"id"`
	}{
		PublicKey:    os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		ClientSecret: pi.ClientSecret,
		ID:           pi.ID,
	})
}

// handleCancelPaymentIntent cancels amount specified in payment intent by specified id.
// https://docs.stripe.com/refunds?dashboard-or-api=api#cancel-payment
func handleCancelPaymentIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	// CaptureRequestParams represents the structure of the request from
	// the client.
	type CancelRequestParams struct {
		PaymentIntentID string `json:"paymentIntentID"`
	}

	// Decode the incoming request
	req := CancelRequestParams{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	params := &stripe.PaymentIntentCancelParams{
		CancellationReason: stripe.String(string(stripe.PaymentIntentCancellationReasonAbandoned)),
	}

	pi, err := paymentintent.Cancel(req.PaymentIntentID, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("paymentintent.Capture: %v", err)
	}

	writeJSON(w, pi)
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("ioutil.ReadAll: %v", err)
		return
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), os.Getenv("STRIPE_WEBHOOK_SECRET"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("webhook.ConstructEvent: %v", err)
		return
	}

	if event.Type == "payment_method.attached" {
		log.Printf("❗ PaymentMethod successfully attached to Customer: %s", event.Data.Raw)
		return
	}

	if event.Type == "payment_intent.succeeded" {
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if string(paymentIntent.SetupFutureUsage) == "" {
			log.Printf("❗ Customer did not want to save the card.")
		}

		log.Printf("💰 Payment received!")
		return
	}

	if event.Type == "payment_intent.payment_failed" {
		log.Printf("❌ Payment failed.")
		return
	}
	if event.Type == "payment_intent.requires_action" {
		log.Printf("💰 Payment requires action: %s", event.Data.Raw)
		return
	}
	if event.Type == "payment_intent.amount_capturable_updated" {
		log.Printf("💰 Payment captured amount updated: %s", event.Data.Raw)
		return
	}
	//if event.Type == "charge.succeeded" {
	//	log.Printf("💰 Charge succeeded: %s", event.Data.Raw)
	//	return
	//}

	writeJSON(w, struct {
		Status string `json:"status"`
	}{
		Status: "success",
	})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewEncoder.Encode: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return
	}
}
