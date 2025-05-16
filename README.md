# ESI SSO

A tool for doing the ESI OAuth and getting the auth and refresh tokens.

## Setup

In the [ESI developer portal](https://developers.eveonline.com/applications)
create an Authentication & API Access application and get the client ID and secret.
The callback URL should be `http://localhost:8080/oauth/callback`.

It should have the following scopes:
* esi-wallet.read_character_wallet.v1
* esi-universe.read_structures.v1
* esi-assets.read_assets.v1
* esi-planets.manage_planets.v1
* esi-markets.read_character_orders.v1
* esi-characters.read_blueprints.v1

[Go](https://go.dev/) is required to run this run this tool. 

## Usage

1. Checkout this repository
2. Get your client id and secret from your ESI application
3. Run the tool with the following command:
```bash
CLIENT_ID=XXXXXXXXX CLIENT_SECRET=XXXXXXX go run main.go
```