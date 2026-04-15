---
name: payments
description: "Payment feature implementation (Stripe). Use when adding subscriptions or one-time payments."
allowed-tools: ["Read", "Write", "Edit", "Bash"]
---

# Payments Skill

A skill for implementing payment features using Stripe.

---

## Features

- Subscriptions (monthly/annual)
- One-time payments
- Webhooks (payment completion notifications)
- Customer portal (plan changes, cancellations)

---

## Execution Flow

1. Check project structure
2. Choose subscription or one-time payment
3. Install Stripe SDK
4. Create payment page
5. Set up webhook endpoint
6. Guide environment variable configuration
