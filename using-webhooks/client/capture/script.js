// A reference to Stripe.js
var stripe;

document.querySelector("#capture").addEventListener("click", function(evt) {
    evt.preventDefault();
    changeLoadingState(true);
    var piID = document.querySelector("#payment-id").value;
    var piAmount = document.querySelector("#payment-amount").value;
    // Initiate payment
    fetch("/capture-payment-intent", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            "paymentIntentID": piID,
            "amount": piAmount*100,
        })
    })
        .then(function (result) {
            return result.json();
        })
        .then(function (result) {
            changeLoadingState(false);
            var paymentIntentJson = JSON.stringify(result, null, 2);
            document.querySelectorAll(".payment-view").forEach(function (view) {
                view.classList.add("hidden");
            });
            document.querySelectorAll(".completed-view").forEach(function (view) {
                view.classList.remove("hidden");
            });
            document.querySelector(".status").textContent =
                result.status === "succeeded" ? "succeeded" : "did not complete";
            document.querySelector("pre").textContent = paymentIntentJson;

        })
});

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
