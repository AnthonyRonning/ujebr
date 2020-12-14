# UJEBR (Uncle Jim's Emergency Bitcoin Recovery)

UJEBR is a piece of software that tries to help alleviate the problem with unintentional access to your Uncle Jim's (or your friend's or your own) bitcoin funds before it is too late. 

This takes advantage of RBF (Replace By Fee) unconfirmed transactions so that you may still have a chance to redirect the funds somewhere else. The replacement transactions will have a higher fee and are non-RBF so the funds will not be redirected afterwards. 

### Motivation

Everyone should be able to know how to quickly recover from a transaction they (or someone else) just sent unintentionally. By setting this up ahead of time and testing it out on a few testnet transactions, you should feel confident to act on this if and when something goes wrong.

Unfortunately I've found no mainstream wallets that I personally use allow replaceable transactions to alternative addresses. I've seen fee bumps and such, but not changing the address. 

Additionally, I would like this project to evolve a bit after the core implementation so that very savvy and trusted bitcoiners can be prepared to take action if a close friend comes to them needing quick help. I don't advocate giving out xpubs to people but in some cases a close friend or relative can't be trusted on their own. But hopefully multisig becomes more mainstream and can help avoid these things from happening so often.

### Limitations

This will not work on transactions that have been originally sent with the non-RBF flag and of course will not work on transactions with a single confirmation.

BWT is pretty fast, but it's best if you have running beforehand with your xpub already passed in. If not, you can set it up to pass in a date to scan from to make it faster. 

### Use cases

  - You made a mistake and sent to the wrong address with RBF on and noticed it before confirmation.
  - An attacker gained access to your private key and sent with a wallet that used RBF by default, you would need to have caught it before the transaction(s) confirm.
  - You are testing out 'double-spent' RBF transactions.


*This software is to be used under emergencies where you or your uncle would otherwise have no other options.*

## Feature list

Very alpha and incomplete but the implemented / expected feature list is below: 

- [x] Scan all unconfirmed RBF transactions (BWT integration).
- [x] Create unsigned transactions redirecting all replaceable funds to desired address.
  - [x] Uses a higher fee and creates the transaction in non-RBF mode.
- [ ] Allow input of a private key, xpriv, or seed phrase to sign replacement transaction.
- [ ] Optionally return unsigned transactions in PSBT format.
- [ ] Integrate BWT so it does not need to be ran separately.
- [ ] Monitor mode when an entire wallet sweep has been made.
- [ ] Simple web-based frontend w/ backend running as a server.
- [ ] Pass in a single transaction id to recover.

## Running

### BWT 

Currently depends on setting up [BWT](https://github.com/shesek/bwt) with the desired XPUB. 

```bash
./bwt -n testnet -x tpub... -a {core_username}:{core_password} 
```

Let BWT complete it's scan, you'll see something like this message: 

```bash
INFO  bwt::indexer > completed initial sync in 51.211344ms up to height 1897523 (total 13 transactions and 11 addresses)
``` 

### UJEBR

```bash
go run backend/cmd/backend/*.go recover --recover_address mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt --bwt.url http://127.0.0.1 --bwt.port 3060
```

Example output (with just unsigned transaction mode): 

```bash
01000000018f5e471ebf89a05c5592ddcebeb5973ebb06376ae091830cd09ca736f1248b630000000000ffffffff01960c0300000000001976a914344a0f48ca150ec2b903817660b9b68b13a6702688ac00000000
0100000001aa75931782a8928de4aabbcb3d384a787e55c3982f5a26e642a19e3e79d46acd0100000000ffffffff015a0c0100000000001976a914344a0f48ca150ec2b903817660b9b68b13a6702688ac00000000
```

If there are no pending replaceable transactions available, then the program will return:

```
panic: no transactions replaceable
```
