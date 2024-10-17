// A reference to Stripe.js
var stripe;

// Information about the order
// Used on the server to calculate order total
var orderData = {
  items: [{ id: "photo-subscription" }],
  currency: "usd"
};

fetch("/create-setup-intent", {
  method: "POST",
  headers: {
    "Content-Type": "application/json"
  },
  body: JSON.stringify(orderData)
})
  .then(function(result) {
    return result.json();
  })
  .then(function(data) {
    return setupElements(data);
  })
  .then(function(stripeData) {
    document.querySelector("#submit").addEventListener("click", function(evt) {
      evt.preventDefault();
      // Initiate payment
      confirm(stripeData.stripe, stripeData.card, stripeData.clientSecret, stripeData.elements);
    });
  });

// Set up Stripe.js and Elements to use in checkout form
var setupElements = function(data) {
  stripe = Stripe(data.publicKey);
  var clientSecret = data.clientSecret;
  var elements = stripe.elements({ clientSecret });
  var payment = elements.create("payment");
  // var card = elements.create("payment", { style: style });
  payment.mount("#payment-element");

  return {
    stripe: stripe,
    card: payment,
    clientSecret: data.clientSecret,
    id: data.id,
    elements: elements
  };
};

/*
 * Calls stripe.confirmCardPayment which creates a pop-up modal to
 * prompt the user to enter  extra authentication details without leaving your page
 */
var confirm = function(stripe, card, clientSecret, elements) {
  changeLoadingState(true);

  // Initiate the payment.
  // If authentication is required, confirmCardPayment will automatically display a modal

  // Use setup_future_usage to save the card and tell Stripe how you plan to charge it in the future
  stripe
    .confirmSetup({
      elements,
      redirect: 'if_required'
    })
    .then(function(result) {
      if (result.error) {
        changeLoadingState(false);
        var errorMsg = document.querySelector(".sr-field-error");
        errorMsg.textContent = result.error.message;
        setTimeout(function() {
          errorMsg.textContent = "";
        }, 4000);
      } else {
        confirmComplete(clientSecret);
        // There's a risk the customer will close the browser window before the callback executes
        // Fulfill any business critical processes async using a 
        // In this sample we use a webhook to listen to payment_intent.succeeded 
        // and add the PaymentMethod to a Customer
      }
    });
};

/* ------- Post-payment helpers ------- */

// Shows a success / error message when the payment is complete
var confirmComplete = function(clientSecret) {
  stripe.retrieveSetupIntent(clientSecret).then(function(result) {
    var setupIntent = result.setupIntent;
    var setupIntentJson = JSON.stringify(setupIntent, null, 2);
    document.querySelectorAll(".payment-view").forEach(function(view) {
      view.classList.add("hidden");
    });
    document.querySelectorAll(".completed-view").forEach(function(view) {
      view.classList.remove("hidden");
    });
    document.querySelector(".status").textContent =
      setupIntent.status === "succeeded" ? "succeeded" : "did not complete";
    document.querySelector("pre").textContent = setupIntentJson;
  });
};

// Show a spinner on payment submission
var changeLoadingState = function(isLoading) {
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
