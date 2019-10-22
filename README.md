# safeway Offers

Safeway is a major US supermarket chain that offer a free members card.
Members can use the Safeway app or site to load offers (either personal offers based on your shopping history, or generic manufacturer coupons).
These offers are only applied if they are "loaded" into your member account.

This utility will use the Safeway mobile api's to get all offers, and load them all into your account.
This way you will not lose any money saving offers.

You can schedule this to run every week to load the new offers

## Usage
##### Windows
- Run `cmd.exe`
- `cd` to the folder containing `safeway-offers.exe`
- Run `safeway-offers.exe -u "<SAFEWAY_USERNAME>" -p "<SAFEWAY_PASSWORD>" -id "<SAFEWAY-SHOP-ID>`
##### macOS or Linux
- Open your Terminal
- `cd` to the folder containing `safeway-offers`
- `chmod +x safeway-offers` to make it executable
- Run `./safeway-offers -u "<SAFEWAY_USERNAME>" -p "<SAFEWAY_PASSWORD>" -id "<SAFEWAY-STORE-ID>` 


## Finding your Safeway Store ID
- Go to https://local.safeway.com/safeway.html
- Find your local store
- Hover over the "weekly Ad" link, the link will appear at the bottom, and be something like
`https://www.safeway.com/set-store.html?storeId=2948&target=weeklyad` 
- The store id is the 4 digits that comes after `storeId=`

## Building
- Install and setup latest Go
- Get the module and its dependencies: `go get -u github.com/giwty/safeway-offers`
- Build it for the OS you need, and make sure to choose `amd64` architecture:
    - `env GOOS=target-OS GOARCH=amd64 go build github.com/giwty/safeway-offers`
    - `target-OS` can be `windows`, `darwin` (mac OS), `linux`, or any other (check the Go documentation for a complete list).
