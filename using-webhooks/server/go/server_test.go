package main

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v80"
	"github.com/stripe/stripe-go/v80/customer"
	"github.com/stripe/stripe-go/v80/paymentintent"
	"github.com/stripe/stripe-go/v80/paymentmethod"
)

func Test_PayAgain(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv.Load: %v", err)
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// The customer ID and payment method ID should be retrieved from your database
	// where you stored them when the Setup Intent was confirmed.
	customerID := "cus_R2wA35BYRGKC8o"
	paymentMethodID := "pm_1QDmpDAJlbf9cOtYb1fZJyw9"

	paymentMethod, err := paymentmethod.Get(paymentMethodID, nil)
	require.NoError(t, err)

	// https://docs.stripe.com/payments/save-during-payment?platform=react-native&mobile-ui=payment-element#react-native-charge-saved-payment-method
	// Create a PaymentIntent to charge the customer
	params := &stripe.PaymentIntentParams{
		Amount:                    stripe.Int64(310), // Amount in cents
		Currency:                  stripe.String(string(stripe.CurrencyUSD)),
		Customer:                  stripe.String(customerID),
		PaymentMethod:             stripe.String(paymentMethodID),
		PaymentMethodTypes:        []*string{stripe.String(string(paymentMethod.Type))},
		Confirm:                   stripe.Bool(true),
		StatementDescriptor:       stripe.String("firebolt"),
		StatementDescriptorSuffix: stripe.String("invoice due"),
		Description:               stripe.String("Invoice #4137591vf"),
		//https://docs.stripe.com/payments/payment-intents/asynchronous-capture
		CaptureMethod: stripe.String("automatic_async"),
		OffSession:    stripe.Bool(true),
	}

	// You can optionally set the Setup Intent ID if you want to reference it
	//params.SetupIntent = stripe.String("seti_123456789")

	pi, err := paymentintent.New(params)
	if err != nil {
		log.Fatalf("Failed to create payment intent: %v", err)
	}
	fmt.Printf("PaymentIntent created: %s\n", pi.ID)
	if pi.Status == stripe.PaymentIntentStatusSucceeded {
		// Handle post-payment fulfillment
		fmt.Printf("PaymentIntent succeeded: %s\n", pi.ID)
	} else if pi.Status == stripe.PaymentIntentStatusRequiresAction {
		// Tell the client to handle the action
		fmt.Printf("PaymentIntent requires action: %s, %s\n", pi.ID, pi.ClientSecret)
	} else {
		fmt.Printf("PaymentIntent status: %s\n", pi.Status)
	}
}

// https://docs.stripe.com/payments/save-during-payment?platform=web#charge-saved-payment-method
func Test_PayAgainWithFallbackToOnSession(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv.Load: %v", err)
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// The customer ID and payment method ID should be retrieved from your database
	// where you stored them when the Setup Intent was confirmed.
	customerID := "cus_R2wA35BYRGKC8o"
	paymentMethodID := "pm_1QAqIeAJlbf9cOtYXOevXdye"

	paymentMethod, err := paymentmethod.Get(paymentMethodID, nil)
	require.NoError(t, err)

	// https://docs.stripe.com/payments/save-during-payment?platform=react-native&mobile-ui=payment-element#react-native-charge-saved-payment-method
	// Create a PaymentIntent to charge the customer
	params := &stripe.PaymentIntentParams{
		Amount:                    stripe.Int64(210), // Amount in cents
		Currency:                  stripe.String(string(stripe.CurrencyUSD)),
		Customer:                  stripe.String(customerID),
		PaymentMethod:             stripe.String(paymentMethodID),
		Confirm:                   stripe.Bool(true),
		StatementDescriptor:       stripe.String("firebolt"),
		StatementDescriptorSuffix: stripe.String("invoice due"),
		Description:               stripe.String("Invoice #4137591vf"),
		//https://docs.stripe.com/payments/payment-intents/asynchronous-capture
		CaptureMethod: stripe.String("automatic_async"),
		OffSession:    stripe.Bool(true),
	}

	// You can optionally set the Setup Intent ID if you want to reference it
	//params.SetupIntent = stripe.String("seti_123456789")

	pi, err := paymentintent.New(params)
	if err != nil {
		sErr, ok := err.(*stripe.Error)
		if ok && sErr.Code == stripe.ErrorCodeAuthenticationRequired {
			log.Printf("PaymentIntent failed as requires authentication, creating new one instead: %s\n", pi.ID)
			// create on session payment intent
			params.OffSession = nil
			params.PaymentMethodTypes = []*string{stripe.String(string(paymentMethod.Type))}

			// You can optionally set the Setup Intent ID if you want to reference it
			//params.SetupIntent = stripe.String("seti_123456789")
			pi, err = paymentintent.New(params)
			if err != nil {
				log.Fatalf("Failed to create payment intent: %v", err)
			}
		} else {
			log.Fatalf("Failed to create payment intent: %v", err)
		}
	}

	fmt.Printf("PaymentIntent created: %s\n", pi.ID)
	if pi.Status == stripe.PaymentIntentStatusSucceeded {
		// Handle post-payment fulfillment
		fmt.Printf("PaymentIntent succeeded: %s\n", pi.ID)
	} else if pi.Status == stripe.PaymentIntentStatusRequiresAction {
		// Tell the client to handle the action
		fmt.Printf("PaymentIntent requires action: %s, %s\n", pi.ID, pi.ClientSecret)
	} else {
		fmt.Printf("PaymentIntent status: %s\n", pi.Status)
	}
}

// TODO need to run on every confirmation failure
func Test_UpdatePaymentIntent(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv.Load: %v", err)
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// The customer ID and payment method ID should be retrieved from your database
	// where you stored them when the Setup Intent was confirmed.
	paymentIntentID := "pi_3QA7SMAJlbf9cOtY0lJlpoyQ"
	paymentMethodID := "pm_1QAqIeAJlbf9cOtYXOevXdye"
	// https://docs.stripe.com/payments/save-during-payment?platform=react-native&mobile-ui=payment-element#react-native-charge-saved-payment-method
	// Create a PaymentIntent to charge the customer
	params := &stripe.PaymentIntentParams{
		Amount:                    stripe.Int64(400), // Amount in cents
		Currency:                  stripe.String(string(stripe.CurrencyUSD)),
		PaymentMethod:             stripe.String(paymentMethodID),
		StatementDescriptor:       stripe.String("firebolt"),
		StatementDescriptorSuffix: stripe.String("invoice due"),
		Description:               stripe.String("Invoice #4137591vf"),
	}

	// You can optionally set the Setup Intent ID if you want to reference it
	//params.SetupIntent = stripe.String("seti_123456789")

	pi, err := paymentintent.Update(paymentIntentID, params)
	if err != nil {
		log.Fatalf("Failed to update payment intent: %v", err)
	}
	fmt.Printf("PaymentIntent created: %s\n", pi.ID)
	if pi.Status == stripe.PaymentIntentStatusSucceeded {
		// Handle post-payment fulfillment
		fmt.Printf("PaymentIntent succeeded: %s\n", pi.ID)
	} else if pi.Status == stripe.PaymentIntentStatusRequiresAction {
		// Tell the client to handle the action
		fmt.Printf("PaymentIntent requires action: %s, %s\n", pi.ID, pi.ClientSecret)
	} else {
		fmt.Printf("PaymentIntent status: %s\n", pi.Status)
	}
}

func Test_PayAgainWithSuccessfulPaymentMethod(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv.Load: %v", err)
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// TODO refactor
	// 1. select default payment method
	// 2.1. try create payment intent with off session
	// 2.2. if failed with requires authentication, create on session payment intent
	// 3. if failed find last successful payment method
	// 4. repeat 2
	// 5. if failed, create blank payment intent for on session confirmation with payment details

	// The customer ID and payment method ID should be retrieved from your database
	// where you stored them when the Setup Intent was confirmed.
	customerID := "cus_R2wA35BYRGKC8o"
	var paymentMethod *stripe.PaymentMethod

	cDetails, err := customer.Get(customerID, &stripe.CustomerParams{
		Expand: []*string{stripe.String("invoice_settings.default_payment_method")},
	})
	// default payment method is in invoice_settings.default_payment_method, or in default_source
	// we should use in our flow invoice_settings.default_payment_method from UI
	require.NoError(t, err)
	paymentMethod = cDetails.InvoiceSettings.DefaultPaymentMethod

	// https://docs.stripe.com/payments/save-during-payment?platform=react-native&mobile-ui=payment-element#react-native-charge-saved-payment-method
	// Create a PaymentIntent to charge the customer
	params := &stripe.PaymentIntentParams{
		Amount:                    stripe.Int64(310), // Amount in cents
		Currency:                  stripe.String(string(stripe.CurrencyUSD)),
		Customer:                  stripe.String(customerID),
		PaymentMethod:             stripe.String(paymentMethod.ID),
		PaymentMethodTypes:        []*string{stripe.String(string(paymentMethod.Type))},
		Confirm:                   stripe.Bool(true),
		StatementDescriptor:       stripe.String("firebolt"),
		StatementDescriptorSuffix: stripe.String("invoice due"),
		Description:               stripe.String("Invoice #4137591vf"),
		//https://docs.stripe.com/payments/payment-intents/asynchronous-capture
		CaptureMethod: stripe.String("automatic_async"),
		OffSession:    stripe.Bool(true),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		sErr, ok := err.(*stripe.Error)
		if ok && sErr.Code == stripe.ErrorCodeAuthenticationRequired {
			log.Printf("PaymentIntent failed as requires authentication, creating new one instead: %s\n", pi.ID)
			// create on session payment intent
			params.OffSession = nil
			params.PaymentMethodTypes = []*string{stripe.String(string(paymentMethod.Type))}
			pi, err = paymentintent.Confirm(sErr.PaymentIntent.ID, &stripe.PaymentIntentConfirmParams{
				PaymentMethod:      stripe.String(paymentMethod.ID),
				PaymentMethodTypes: []*string{stripe.String(string(paymentMethod.Type))},
				//https://docs.stripe.com/payments/payment-intents/asynchronous-capture
				CaptureMethod: stripe.String("automatic_async"),
				OffSession:    stripe.Bool(false),
			})
			if err != nil {
				log.Fatalf("Failed to confirm payment intent: %v", err)
			}
		} else {
			log.Fatalf("Failed to create payment intent: %v", err)
			// TODO fallback to last successful payment method
			//piList := paymentintent.List(&stripe.PaymentIntentListParams{
			//	Customer: stripe.String(customerID),
			//	Expand:   []*string{stripe.String("data.payment_method")},
			//})
			//require.NoError(t, piList.Err())
			//
			//for piList.Next() {
			//	pi := piList.PaymentIntent()
			//	if pi.Status == stripe.PaymentIntentStatusSucceeded {
			//		paymentMethod = pi.PaymentMethod
			//		break
			//	}
			//}
		}
	}

	fmt.Printf("PaymentIntent created: %s\n", pi.ID)
	if pi.Status == stripe.PaymentIntentStatusSucceeded {
		// Handle post-payment fulfillment
		fmt.Printf("PaymentIntent succeeded: %s\n", pi.ID)
	} else if pi.Status == stripe.PaymentIntentStatusRequiresAction {
		// Tell the client to handle the action
		fmt.Printf("PaymentIntent requires action: %s, %s\n", pi.ID, pi.ClientSecret)
	} else {
		fmt.Printf("PaymentIntent status: %s\n", pi.Status)
	}
}

func TestRedisplayCheck(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv.Load: %v", err)
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	//pm_1QAskkAJlbf9cOtYK9N9Gcqz

	//paymentMethods := customer.ListPaymentMethods(&stripe.CustomerListPaymentMethodsParams{
	//	Customer: stripe.String(customerID),
	//})
	//require.NoError(t, paymentMethods.Err())
	//for paymentMethods.Next() {
	//	pm := paymentMethods.PaymentMethod()
	//	t.Logf("PaymentMethod: %s, %s\n", pm.ID, pm.Type)
	//}
	//customerID := "cus_R2wA35BYRGKC8o"

	pm, err := paymentmethod.Update("pm_1QAskkAJlbf9cOtYK9N9Gcqz", &stripe.PaymentMethodParams{
		AllowRedisplay: stripe.String(string(stripe.PaymentMethodAllowRedisplayAlways)),
	})
	require.NoError(t, err)
	t.Logf("PaymentMethod: %s, %s\n", pm.ID, pm.Type)
}
