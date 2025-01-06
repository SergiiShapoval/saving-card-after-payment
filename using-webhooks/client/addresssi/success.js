document.addEventListener('DOMContentLoaded', async (e) => {
  // Initialize Stripe.js with your publishable key.
  const {publishableKey} = await fetch("/config").then(res => res.json());
  const stripe = Stripe(publishableKey);

  // Get the PaymentIntent clientSecret from query string params.
  const params = new URLSearchParams(window.location.search);
  const clientSecret = params.get('setup_intent_client_secret');

  // Retrieve the PaymentIntent.
  const {setupIntent} = await stripe.retrieveSetupIntent(clientSecret)
  addMessage("Setup Intent Status: " + setupIntent.status);
  addMessage(setupIntent.id);
});
