LNbits Infinity (BETA)
======

![Lightning network wallet](https://i.imgur.com/EHvK6Lq.png)

# Free Open-Source Lightning Accounts System with Extensions

For the Python version of LNbits checkout <a href="https://github.com/lnbits/lnbits-legend/">Legend</a>. Demo servers available on [lnbits.com](https://lnbits.com).

Join us on [https://t.me/lnbits](https://t.me/lnbits).

LNbits can run on top of any lightning-network funding source, currently there is support for LND, c-lightning, Spark, LNpay, OpenNode, lntxbot, with more being added regularly.

Checkout [Awesome-LNbits](https://github.com/cryptoteun/awesome-lnbits), a currated list of projects made using LNbits.

Checkout the LNbits [YouTube](https://www.youtube.com/playlist?list=PLPj3KCksGbSYG0ciIQUWJru1dWstPHshe) video series.

### Required dependencies

```
sudo apt-get install lua5.3
go install github.com/joho/godotenv/cmd/godotenv@latest
```

### Running LNbits in development mode

Install [Air](https://github.com/cosmtrek/air).

Open two terminals. In one, do

```
cd client
quasar dev
```

on the other, do

```
QUASAR_DEV_SERVER=http://localhost:8080 make dev
```
