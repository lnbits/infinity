LNbits Infinity (BETA)
======

### Installation

```sh
git clone https://github.com/lnbits/infinity.git

# install dependencies
sudo apt-get install lua5.3-dev
```

Move into the Infinity directory:

```sh
cd infinity
```

Setup environment variables:

```sh
PORT=6000
QUASAR_DEV_SERVER=http://localhost:6001

DATABASE=dev.sqlite

LIGHTNING_BACKEND=void # adjust accordingly
# depending on the lightning backend chosen you'll need different environment variables
# check https://github.com/lnbits/infinity/blob/77dafa306b0ea79cdf5ffa9bf39c13cd04ffdfa6/lightning/backend.go#L18-L28 file for more information

SITE_TITLE=My Infinity
SITE_TAGLINE=An infinitude of wallets and apps
SITE_DESCRIPTION=Local server of LNbits Infinity
DEFAULT_WALLET_NAME=Wallet

SECRET=typesomethingrandomthisisnotsuperimportantjustalittle
```

Install [Air](https://github.com/cosmtrek/air).

Finally, open two terminals. In one, do

```sh
cd client
./node_modules/.bin/quasar dev -p 6001
```

on the other, do

```sh
make dev
```

Then access your LNbits Infinity at http://localhost:6000/ (not :6001).
