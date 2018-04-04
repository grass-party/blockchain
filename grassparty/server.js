'use strict';

var Fabric_Client = require('fabric-client');
var path = require('path');
var util = require('util');
var os = require('os');


var express = require("express");
var bodyParser = require('body-parser');
var app = express();
var port = 3000;

app.use(bodyParser.urlencoded({extended:false}));
app.use(bodyParser.json());

app.post("/register",function(req, res){
	console.log("register");

	var account_id = req.body.account_id;
	var account_pubkey = req.body.account_pubkey;
	var request = {
		chaincodeId: 'grassparty',
		fcn: 'register',
		args: [account_id, account_pubkey],
		chainId: 'mychannel',
		};

	restAPI_Post(request, function(result){
		if(result[0] == 'Success'){
			res.send("user registered : " + result[1]);
		} else {
			res.send("user register fail : " + result[1]);
		}
	});
});

app.post("/setAgenda",function(req, res){
	console.log("setAgenda");

	var agenda_id = req.body.agenda_id;
	var agenda_hash = req.body.agenda_hash;

	var request = {
		chaincodeId: 'grassparty',
		fcn: 'setAgenda',
		args: [agenda_id, agenda_hash],
		chainId: 'mychannel',
	};

	restAPI_Post(request, function(result){
		if(result[0] == 'Success'){
			res.send("set Agenda : " + result[1]);
		} else {
			res.send("set Agenda fail : " + result[1]);
		}
	});
});

app.post("/vote",function(req, res){
	console.log("vote");

/*	var account_id = req.body.account_id;
	var agenda_id = req.body.agenda_id;
	var vote = req.body.vote;
	var sign = req.body.sign;

       	var request = {
               	chaincodeId: 'grassparty',
               	fcn: 'vote',
               	args: [account_id, agenda_id, vote, sign],
               	chainId: 'mychannel',
       	};
*/

/*
	var vote_data_json = req.body.vote_data_json;

       	var request = {
               	chaincodeId: 'grassparty',
               	fcn: 'vote',
               	args: [vote_data_json],
               	chainId: 'mychannel',
       	};
*/

	var id = req.body.id;
	var ids = id.split("-");
	var account_id = ids[0];
	var agenda_id = ids[1];
	var vote = "";
	switch(req.body.data) {
		case "1":
			vote = "A";
			break;
		case "2":
			vote = "B";
			break;
		case "3":
			vote = "C";
			break;
		case "4":
			vote = "D";
			break;
		case "5":
			vote = "E";
			break;
		default:
			vote = "F";
	}

	console.log("vote:"+vote)

	var sign = "asdf";

       	var request = {
               	chaincodeId: 'grassparty',
               	fcn: 'vote',
               	args: [account_id, agenda_id, vote, sign],
               	chainId: 'mychannel',
       	};

	restAPI_Post(request, function(result){
		if(result[0] == 'Success'){
			res.send("user voted : " + result[1]);
		} else {
			res.send("user vote fail : " + result[1]);
		}
	});
});

app.get("/getAgenda",function(req, res){
	console.log("getAgenda");

	var agenda_id = req.query.agenda_id;

	console.log("agenda_id");
	console.log(agenda_id);

	var request = {
		chaincodeId: 'grassparty',
		fcn: 'getAgenda',
		args: [agenda_id]
	};

	restAPI_Get(request, function(result){
		if(result[0] == 'Success'){
			res.send(result[1]);
		} else {
			res.send(result[1]);
		}
	});
});


app.get("/getSymmetricKey",function(req, res){
	console.log("getSymmetricKey");

	var account_id = req.query.account_id;

	var request = {
		chaincodeId: 'grassparty',
		fcn: 'getSymmetricKey',
		args: [account_id]
	};

	restAPI_Get(request, function(result){
		if(result[0] == 'Success'){
			res.send(result[1]);
		} else {
			res.send(result[1]);
		}
	});
});


function restAPI_Post(request, callback){

	var fabric_client = new Fabric_Client();

	// setup the fabric network
	var channel = fabric_client.newChannel('mychannel');
	var peer = fabric_client.newPeer('grpc://localhost:7051');
	channel.addPeer(peer);
	var order = fabric_client.newOrderer('grpc://localhost:7050');
	channel.addOrderer(order);

	//
	var member_user = null;
	var store_path = path.join(__dirname, 'hfc-key-store');
	console.log('Store path:'+store_path);
	var tx_id = null;

	var return_str = "";
	var result = [];

	// create the key value store as defined in the fabric-client/config/default.json 'key-value-store' setting
	Fabric_Client.newDefaultKeyValueStore({ path: store_path
	}).then((state_store) => {
		// assign the store to the fabric client
		fabric_client.setStateStore(state_store);
		var crypto_suite = Fabric_Client.newCryptoSuite();
		// use the same location for the state store (where the users' certificate are kept)
		// and the crypto store (where the users' keys are kept)
		var crypto_store = Fabric_Client.newCryptoKeyStore({path: store_path});
		crypto_suite.setCryptoKeyStore(crypto_store);
		fabric_client.setCryptoSuite(crypto_suite);

		// get the enrolled user from persistence, this user will sign all requests
		return fabric_client.getUserContext('user1', true);
	}).then((user_from_store) => {
		if (user_from_store && user_from_store.isEnrolled()) {
			console.log('Successfully loaded user1 from persistence');
			member_user = user_from_store;
		} else {
			throw new Error('Failed to get user1.... run registerUser.js');
		}

		// get a transaction id object based on the current user assigned to fabric client
		tx_id = fabric_client.newTransactionID();
		console.log("Assigning transaction_id: ", tx_id._transaction_id);

		// must send the proposal to endorsing peers
		var req = {
			chaincodeId: request.chaincodeId,
			fcn: request.fcn,
			args: request.args,
			chainId: request.chainId,
			txId: tx_id
		};

		// send the transaction proposal to the peers
		return channel.sendTransactionProposal(req);
	}).then((results) => {
		console.log("results[0]:"+results[0]);
		var proposalResponses = results[0];
		var proposal = results[1];
		let isProposalGood = false;
		if (proposalResponses &&
			proposalResponses[0].response &&
			proposalResponses[0].response.status === 200) {
			isProposalGood = true;
			console.log('Transaction proposal was good');
		} else {
			result[0] = "Fail";
			result[1] = results[0];
			console.error('Transaction proposal was bad');
		}
		if (isProposalGood) {
			console.log(util.format(
			'Successfully sent Proposal and received ProposalResponse: Status - %s, message - "%s"',
			proposalResponses[0].response.status, proposalResponses[0].response.message));

			// build up the request for the orderer to have the transaction committed
			var request = {
				proposalResponses: proposalResponses,
				proposal: proposal
			};

			// set the transaction listener and set a timeout of 30 sec
			// if the transaction did not get committed within the timeout period,
			// report a TIMEOUT status
			var transaction_id_string = tx_id.getTransactionID(); //Get the transaction ID string to be used by the event processing
			var promises = [];

			var sendPromise = channel.sendTransaction(request);
			promises.push(sendPromise); //we want the send transaction first, so that we know where to check status

			// get an eventhub once the fabric client has a user assigned. The user
			// is required bacause the event registration must be signed
			let event_hub = fabric_client.newEventHub();
			event_hub.setPeerAddr('grpc://localhost:7053');

			// using resolve the promise so that result status may be processed
			// under the then clause rather than having the catch clause process
			// the status
			let txPromise = new Promise((resolve, reject) => {
				let handle = setTimeout(() => {
					event_hub.disconnect();
					resolve({event_status : 'TIMEOUT'}); //we could use reject(new Error('Trnasaction did not complete within 30 seconds'));
					}, 3000);
				event_hub.connect();
				event_hub.registerTxEvent(transaction_id_string, (tx, code) => {
					// this is the callback for transaction event status
					// first some clean up of event listener
					clearTimeout(handle);
					event_hub.unregisterTxEvent(transaction_id_string);
					event_hub.disconnect();

					// now let the application know what happened
					var return_status = {event_status : code, tx_id : transaction_id_string};
					if (code !== 'VALID') {
						console.error('The transaction was invalid, code = ' + code);
						resolve(return_status); // we could use reject(new Error('Problem with the tranaction, event status ::'+code));
					} else {
						console.log('The transaction has been committed on peer ' + event_hub._ep._endpoint.addr);
						resolve(return_status);
					}
				}, (err) => {
					//this is the callback if something goes wrong with the event registration or processing
					reject(new Error('There was a problem with the eventhub ::'+err));
				});
			});
			promises.push(txPromise);

			return Promise.all(promises);
		} else {
			console.error('Failed to send Proposal or receive valid response. Response null or status is not 200. exiting...');
			throw new Error('Failed to send Proposal or receive valid response. Response null or status is not 200. exiting...');
		}
	}).then((results) => {
		console.log('Send transaction promise and event listener promise have completed');
		// check the results in the order the promises were added to the promise all list
		if (results && results[0] && results[0].status === 'SUCCESS') {
			console.log('Successfully sent transaction to the orderer.');
		} else {
			console.error('Failed to order the transaction. Error code: ' + response.status);
		}

		if(results && results[1] && results[1].event_status === 'VALID') {
			console.log('Successfully committed the change to the ledger by the peer');
			result[0] = 'Success';
			result[1] = 'Success';
			callback(result);
		} else {
			console.log('Transaction failed to be committed to the ledger due to ::'+results[1].event_status);
		}
	}).catch((err) => {
		console.error('Failed to invoke successfully :: ' + err);

		callback(result);
	});
}

function restAPI_Get(request, callback){
	var fabric_client = new Fabric_Client();

	// setup the fabric network
	var channel = fabric_client.newChannel('mychannel');
	var peer = fabric_client.newPeer('grpc://localhost:7051');
	channel.addPeer(peer);

	//
	var member_user = null;
	var store_path = path.join(__dirname, 'hfc-key-store');
	console.log('Store path:'+store_path);
	var tx_id = null;


	var result = [];
	// create the key value store as defined in the fabric-client/config/default.json 'key-value-store' setting
	Fabric_Client.newDefaultKeyValueStore({ path: store_path
	}).then((state_store) => {
		// assign the store to the fabric client
		fabric_client.setStateStore(state_store);
		var crypto_suite = Fabric_Client.newCryptoSuite();
		// use the same location for the state store (where the users' certificate are kept)
		// and the crypto store (where the users' keys are kept)
		var crypto_store = Fabric_Client.newCryptoKeyStore({path: store_path});
		crypto_suite.setCryptoKeyStore(crypto_store);
		fabric_client.setCryptoSuite(crypto_suite);

		// get the enrolled user from persistence, this user will sign all requests
		return fabric_client.getUserContext('user1', true);
	}).then((user_from_store) => {
		if (user_from_store && user_from_store.isEnrolled()) {
			console.log('Successfully loaded user1 from persistence');
			member_user = user_from_store;
		} else {
			throw new Error('Failed to get user1.... run registerUser.js');
		}

		// queryCar chaincode function - requires 1 argument, ex: args: ['CAR4'],
		// queryAllCars chaincode function - requires no arguments , ex: args: [''],

		// send the query proposal to the peer
		return channel.queryByChaincode(request);
	}).then((query_responses) => {
		console.log("Query has completed, checking results");
		// query_responses could have more than one  results if there multiple peers were used as targets
		if (query_responses && query_responses.length == 1) {
			if (query_responses[0] instanceof Error) {
				console.error("error from query = ", query_responses[0]);
				result[0] = "Fail";
				result[1] = query_responses[0].toString();

				callback(result);
			} else {
//				res.send(query_responses[0].toString());
				result[0] = "Success";
				result[1] = query_responses[0].toString();
				console.log("Response is ", query_responses[0].toString());

				callback(result);
			}
		} else {
			console.log("No payloads were returned from query");
		}
	}).catch((err) => {
		console.error('Failed to query successfully :: ' + err);
	});
}

app.listen(port, function(){
	console.log("Server started on port %d", port);
});

