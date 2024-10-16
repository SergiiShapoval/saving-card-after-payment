// A reference to Stripe.js
var stripe;

document.querySelector("#confirm").addEventListener("click", function(evt) {
    evt.preventDefault();
    var piID = document.querySelector("#payment-id").value;
    // Initiate payment
    fetch("/confirm-payment-intent", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            "paymentIntentID": piID
        })
    })
        .then(function (result) {
            return result.json();
        })
        .then(function (data) {
            return setupElements(data);
        })
        .then(function (stripeData) {
            confirm(stripeData.stripe, stripeData.clientSecret);
        });
});



// Set up Stripe.js and Elements to use in checkout form
var setupElements = function (data) {
    stripe = Stripe(data.publicKey);

    return {
        stripe: stripe,
        clientSecret: data.clientSecret,
        id: data.id
    };
};

/*
 * Calls stripe.confirmCardPayment which creates a pop-up modal to
 * prompt the user to enter  extra authentication details without leaving your page
 */
var confirm = function (stripe, clientSecret) {

    changeLoadingState(true);

    // Initiate the payment.
    // If authentication is required, confirmCardPayment will automatically display a modal


    // Initiate payment
    stripe.confirmCardPayment(clientSecret, {
        setup_future_usage: "off_session"
    }).then(function (result) {
        const {error: errorAction, paymentIntent} = result;
        changeLoadingState(false);
        if (errorAction) {
            // Show error from Stripe.js in payment form
            console.log("handle card error action: ", errorAction);
            var errorMsg = document.querySelector(".sr-field-error");
            errorMsg.textContent = result.error.message;
            setTimeout(function () {
                errorMsg.textContent = "";
            }, 4000);
        } else {
            // The card action has been handled
            // The PaymentIntent can be confirmed again on the server
            var paymentIntentJson = JSON.stringify(paymentIntent, null, 2);
            document.querySelectorAll(".payment-view").forEach(function (view) {
                view.classList.add("hidden");
            });
            document.querySelectorAll(".completed-view").forEach(function (view) {
                view.classList.remove("hidden");
            });
            document.querySelector(".status").textContent =
                paymentIntent.status === "succeeded" ? "succeeded" : "did not complete";
            document.querySelector("pre").textContent = paymentIntentJson;
        }
        // Handle the result here
    })
        .catch(function (error) {
            console.log("Error on handle card action: ", error);
            changeLoadingState(false);
            var errorMsg = document.querySelector(".sr-field-error");
            errorMsg.textContent = error.message;
            setTimeout(function () {
                errorMsg.textContent = "";
            }, 4000);
        });

};

/* ------- Post-payment helpers ------- */

// Show a spinner on payment submission
var changeLoadingState = function (isLoading) {
    if (isLoading) {
        document.querySelector("button").disabled = true;
        document.querySelector("#spinner").classList.remove("hidden");
        document.querySelector("#button-text").classList.add("hidden");
    } else {
        document.querySelector("button").disabled = false;
        document.querySelector("#spinner").classList.add("hidden");
        document.querySelector("#button-text").classList.remove("hidden");
    }
};
