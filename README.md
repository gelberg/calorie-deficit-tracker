# What is it?

This is a console application that tracks energy deficit basing on data from FatSecret API (calories consumption) and Google Fit API (calories expenditure). It is written in Go and consists of 3 microservices:
- FatSecret producer - periodically polls FatSecret API for today consumption and sends it to Kafka
- Google Fit producer - periodically polls Google Fit API for today expenditure and sends it to Kafka
- Calorie deficit calculator - calulcates calorie deficit basing on consumption and expenditure received through Kafka

  # How to use?

1. Obtain credentials on FatSecret (REST API OAuth 1.0 Credentials) and Google Fit (OAuth 2.0) platforms **or just ask me for them**
2. Clone this repository
3. Enter obtained credentials in `fatsecret_producer/client.json` and `google_fit_producer/client.json`
4. Run `docker compose up` from project directory
5. Attach to `calorie-deficit-tracker-fatsecret-producer` container and follow printed URL - you'll be redirected to FatSecret authorization form
6. On successful authorization OAuth 1.0 verification code is printed - enter it in stdin of an attached container
7. Attach to `calorie-deficit-tracker-google-fit-producer` container and follow printed URL - you'll be redirected to Google authorization form
8. On successful authorization you will be redirected to Google OAuth 2.0 Playground page. Copy Authorization code and enter it in stdin of an attached container
9. Attach to or print logs of `calorie-deficit-tracker-calorie-deficit-calculator` - it periodically prints calorie deficit (i.e. difference between calorie expenditure and consumption) 
