/**
 * Copyright 2017 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// This is an end-to-end test that focuses on exercising all parts of the fabric APIs
// in a happy-path scenario
'use strict';

let config = require("config");
let CONFIG =config['asset'];
let helper = require('../helper.js');
let utils = require('fabric-client/lib/utils.js');
let logger = utils.getLogger('blockchainService');

//for block api
const ADMIN_USER = CONFIG['adminUser'];


let BlockchainService = function() {
    let vm = this;

    vm.install = function(org,chaincode_id, chaincode_path, version,language,get_admin) {
        // logger.info('install called. path:' + chaincode_path + ' id: ' + chaincode_id + ' Ver: ' + version);


        let client;
        let targets;

        return new Promise(function(resolve, reject){

            helper.initObject(ADMIN_USER, org,true).then((clientObj)=>{
                client	 = clientObj.client;
                targets = clientObj.targets;

                // return helper.getSubmitter(client,get_admin, org);
                let username = 'peer'+org+'Admin';
                return client.getUserContext(username);
            }).then((submitter)=>{
                if(submitter){
                    const request = {
                        targets: targets,
                        chaincodePath: chaincode_path,
                        chaincodeId: chaincode_id,
                        chaincodeType: language,
                        chaincodeVersion: version
                    };

                    // send install to endorser
                    return client.installChaincode(request);
                }else{
                    let emsg = "install ERROR : ***err : Admin need enroll at first.";
                    logger.error(emsg);
                    throw new Error(emsg);
                }

            }, (err) => {
                logger.error('Failed to enroll admin. ' + err);
                throw new Error('Failed to enroll admin. ' + err);
            }).then((results) => {
                logger.info('Success installChaincode');
                let proposalResponses = results[0];

                var errMsg = [];
                if (validateProposal(proposalResponses, targets.length, errMsg)) {
                    logger.info('Successfully sent Proposal and received ProposalResponse: Status - %s, message - "%s", metadata - "%s"',
                        proposalResponses[0].response.status, proposalResponses[0].response.message, proposalResponses[0].response.payload);
                    resolve({"result":true});
                } else {
                    let emsg = 'Could NOT confirm all proposal response as endorsement policy. Msg : ' + errMsg[0];
                    logger.error(emsg);
                    throw new Error(emsg);
                }

            }, (err) => {
                let emsg = 'Failed to install due to error: ' + err.stack ? err.stack : err;
                logger.error(emsg);
                reject(err);
            })
        });

    };

    vm.instantiate = function(org, upgrade, chaincodeId, chaincodeVersion, initfnc, args, get_admin) {
        // logger.info('instantiate called. id: ' +  chaincodeId + ' Ver: ' + chaincodeVersion);
        let client;
        let channel;
        let targets;
        let eventhubs;
        let tx_id;

        return new Promise(function(resolve, reject){

            helper.initObject(ADMIN_USER, org,true,'org').then((clientObj)=>{
                client	 = clientObj.client;
                channel	 = clientObj.channel;
                targets = clientObj.targets;
                eventhubs = clientObj.eventhubs;
                // return helper.getSubmitter(client,get_admin, org);
                let username = 'peer'+org+'Admin';
                return client.getUserContext(username);
            }).then((submitter)=>{
                if(submitter){
                    // read the config block from the orderer for the channel
                    // and initialize the verify MSPs based on the participating
                    // organizations
                    return channel.initialize();
                }else{
                    let emsg = "install ERROR : ***err : Admin need enroll at first.";
                    logger.error(emsg);
                    throw new Error(emsg);
                }

            }, (err) => {
                logger.error('Failed to enroll admin. ' + err);
                throw new Error('Failed to enroll admin. ' + err);
            }).then(() => {
                //success then instantiate
                //endorsement policy is used on config.
                tx_id = client.newTransactionID();
                let instantiateRequest = {
                    chaincodeId: chaincodeId,
                    chaincodeVersion: chaincodeVersion,
                    fcn: initfnc,
                    args: args,
                    txId: tx_id,
                    'endorsement-policy': CONFIG.chaincode.endorsement
                };
                //upgrade for same chaincode id instantiation
                if(upgrade){
                    logger.info('** update Chaincode');
                    return channel.sendUpgradeProposal(instantiateRequest, 120000)
                }else{
                    logger.info('** instantiate Chaincode');
                    return channel.sendInstantiateProposal(instantiateRequest, 120000);
                }
            })
                .then((instResults) => {
                    logger.info('Success instantiate/update Chaincode Proposal');
                    let proposalResponses = instResults[0];

                    let errMsg = [];
                    if (validateProposal(proposalResponses, targets.length, errMsg)) {
                        logger.info('Successfully sent Proposal and received ProposalResponse: Status - %s, message - "%s", metadata - "%s"',
                            proposalResponses[0].response.status, proposalResponses[0].response.message, proposalResponses[0].response.payload);
                        //success
                        let request = {
                            proposalResponses: proposalResponses,
                            proposal: instResults[1]
                        };
                        let transactionID = tx_id.getTransactionID();
                        let eventPromises = [];

                        let sendPromise = channel.sendTransaction(request);

                        eventhubs.forEach((eh) => {
                            eh.connect();
                            let txPromise = new Promise((resolve, reject) => {
                                let handle = setTimeout(reject, 30000);
                                eh.registerTxEvent(transactionID.toString(), (tx, code) => {
                                    clearTimeout(handle);
                                    eh.unregisterTxEvent(transactionID.toString());

                                    if (code !== 'VALID') {
                                        logger.info('Transaction was invalid, code = ' + code);
                                        eventhubs.forEach((eh) => {eh.disconnect()});
                                        reject({
                                            errorcode: code,
                                            request : request
                                        });
                                    } else {
                                        // logger.info('Transaction has been committed on peer '+ eh.getPeerAddr());
                                        resolve();
                                    }
                                });
                            });
                            eventPromises.push(txPromise);
                        });

                        return Promise.all([sendPromise].concat(eventPromises))
                            .then((results) => {
                                logger.info('Event promise all complete.');
                                return results[0];

                            }).catch((err) => {
                                eventhubs.forEach((eh) => {eh.disconnect()});
                                throw err;
                            });
                    } else {
                        var emsg = 'Could NOT confirm all proposal response as endorsement policy. Msg : ' + errMsg[0];
                        logger.error(emsg);
                        throw new Error(emsg);
                    }
                }).then((response) => {
                eventhubs.forEach((eh) => {eh.disconnect()});
                if(!response){
                    logger.info("Detected undefined response from eventhub.");
                    resolve({"result":true});
                }else if (response.status && response.status === 'SUCCESS') {
                    logger.info('Successfully sent transaction to the orderer.');
                    logger.info('******************************************************************');
                    logger.info('THIS_TX_IS is : ',tx_id);
                    logger.info('******************************************************************');
                    resolve({"result":true});
                } else {
                    let emsg = 'Failed to order the transaction. Error code: ' + response.status
                    logger.info(emsg);
                    throw new Error(emsg);
                }
            }).catch((err)=>{
                eventhubs.forEach((eh) => {eh.disconnect()});
                err.message +=  "[Error in instantiate/update] " + " [txid is : " + tx_id +  "] " ;
                reject(err);
            });
        });

    };

    vm.invoke = function(enrollId, fnc, args,get_admin) {
        let client;
        let channel;
        let targets;
        let eventhubs;
        let tx_id;
        let org = CONFIG.users[enrollId].org;
        return new Promise(function(resolve, reject){
            helper.initObject(enrollId, org,get_admin,'org').then((clientObj)=>{
                client	 = clientObj.client;
                channel	 = clientObj.channel;
                targets = clientObj.targets;
                eventhubs = clientObj.eventhubs;

                // return helper.getSubmitter(client,get_admin, org, enrollId);
                let username
                if(get_admin){
                    username = 'peer'+org+'Admin';
                }else {
                    username =enrollId;
                }
                return client.getUserContext(username);
            }).then((submitter)=>{
                if(submitter){
                    // read the config block from the orderer for the channel
                    // and initialize the verify MSPs based on the participating
                    // organizations
                    return channel.initialize();
                }else{
                    let emsg = "install ERROR : ***err : Admin need enroll at first.";
                    logger.error(emsg);
                    throw new Error(emsg);
                }

            }, (err) => {
                logger.error('Failed to enroll admin. ' + err);
                throw new Error('Failed to enroll admin. ' + err);
            }).then(() => {
                //transaction id should be generated after setContext.
                // ??? ????????? ???????????? ID??? ????????? ????????? ??????????????? ?????????, ????????? ?????? ????????? ??????
                tx_id   = client.newTransactionID();
                let request = {
                    chaincodeId : CONFIG.chaincode.id,
                    fcn: fnc,
                    args: args,
                    txId: tx_id,
                    targets: targets
                };

                // send proposal to endorser
                // ??? ??????????????? ?????? ????????? ??????
                // logger.info('sendTransactionProposal REQUEST:' + JSON.stringify(request));
                return channel.sendTransactionProposal(request);


            }, (err) => {
                logger.error('Failed to enroll user. ' + err);
                throw new Error('Failed to enroll user. ' + err);
            }).then((results) => {
                // logger.info('Success sendTransactionProposal');
                let proposalResponses = results[0];
                let proposal = results[1];
                let header   = results[2];

                var errMsg = [];
                // ??? ?????? ?????? ??????(Proposal)??? ????????? ?????? - validateProposal ?????? ??????
                if (validateProposal(proposalResponses, targets.length, errMsg)) {
                    logger.info('Successfully sent Proposal and received ProposalResponse: Status - %s, message - "%s", metadata - "%s"',
                        proposalResponses[0].response.status, proposalResponses[0].response.message, proposalResponses[0].response.payload);

                    let request = {
                        proposalResponses: proposalResponses,
                        proposal: proposal,
                        header: header
                    };

                    // set the transaction listener and set a timeout of 30sec for each eventhub
                    // Fail if the transaction did not get committed within the timeout period.
                    let transactionID = tx_id.getTransactionID();
                    let eventPromises = [];
                    // ??? ????????????????????? ????????? ??????????????? Proposal ?????? ?????? ??????
                    let sendPromise = channel.sendTransaction(request);

                    eventhubs.forEach((eh) => {
                        eh.connect();

                        let txPromise = new Promise((resolve, reject) => {
                            let handle = setTimeout(reject, 30000);
                            // ??? ????????? ?????????(EventHub)??? ????????? ???????????? ID ????????? ??????
                            eh.registerTxEvent(transactionID.toString(), (tx, code) => {
                                clearTimeout(handle);
                                eh.unregisterTxEvent(transactionID.toString());
                                // ??? ??????????????? ?????????????????? ?????? ?????? ??????
                                if (code !== 'VALID') {
                                    // logger.info('Transaction was invalid, code = ' + code);
                                    eventhubs.forEach((eh) => {eh.disconnect()});
                                    reject({
                                        errorcode: code,
                                        request : request
                                    });
                                } else {
                                    // logger.info('Transaction has been committed on peer '+ eh.getPeerAddr());
                                    resolve();
                                }
                            });
                        });
                        eventPromises.push(txPromise);
                    });
                    // ?????? ????????? ?????? ????????? ????????? ??? ??????
                    return Promise.all([sendPromise].concat(eventPromises))
                        .then((results) => {
                            // logger.info('Event promise all complete.');
                            return results[0];

                        }).catch((err) => {
                            if(!err){
                                logger.error("Detected undefined error.");
                            }else{
                                // ??? ?????? ?????? ????????? ????????? ?????? ??????????????? ????????? ???????????? ?????? ????????? ??????
                                if(err.errorcode && (err.errorcode === 'MVCC_READ_CONFLICT')){
                                    //user can retry if he/she wants with this errcode.
                                    logger.error("--------------MVCC_READ_CONFLICT-------------------");
                                    throw err;
                                }else{
                                    eventhubs.forEach((eh) => {eh.disconnect()});
                                    throw err;
                                }
                            }
                        });

                } else {
                    let emsg = 'Could NOT confirm all proposal response as endorsement policy. Msg : ' + errMsg[0];
                    logger.error(emsg);
                    throw new Error(emsg);
                }

            }, (err) => {
                let emsg = 'Failed to send proposal due to error: ' + err.stack ? err.stack : err;
                logger.error(emsg);
                throw new Error(emsg);
            }).then((response) => {
                // ??? EventHub?????? ????????? ?????? ?????? ??????
                eventhubs.forEach((eh) => {eh.disconnect()});
                if(!response){
                    // logger.info("Detected undefined response from eventhub.");
                    resolve({"result":true});
                }else if (response.status && response.status === 'SUCCESS') {
                    logger.info('Successfully sent transaction to the orderer.');
                    logger.info('******************************************************************');
                    logger.info('THIS_TX_IS is : ',tx_id);
                    logger.info('******************************************************************');
                    resolve({"result":true});
                } else {
                    let emsg = 'Failed to order the transaction. Error code: ' + response.status
                    // logger.info(emsg);
                    throw new Error(emsg);
                }
            }).catch((err)=>{
                eventhubs.forEach((eh) => {eh.disconnect()});
                err.message +=  "[Error in func : " + fnc + "] " + " [txid is : " + tx_id +  "] " ;
                reject(err);
            });
        });

    };


    /**
     * validate ProposalResponse for Endorsement Policy
     */
    let validateProposal = (proposalResponses, minCount, errMsg) => {
        if(!proposalResponses && !proposalResponses.length){
            logger.error('transaction proposal was null');
            return false;
        }
        //check as endorsement-policy
        var count = 0;
        for(let i in proposalResponses) {
            if (proposalResponses[i].response && proposalResponses[i].response.status === 200) {
                logger.info('transaction proposal : No. %s was good', i);
                count++;
            } else {
                logger.info('transaction proposal : No. %s was bad' , i);
                errMsg.push(proposalResponses[i]);
            }
        }

        return count >= minCount;
    };

    vm.register = function(username, secret){
        return new Promise(function(resolve, reject){
            helper.register(username, secret)
                .then((secret)=>{
                    resolve({"result":secret});
                }).catch((err)=>{
                reject(err);
            });
        });
    };


    vm.query = function(enrollId, fnc, args) {
        let client;
        let channel;
        let targets;
        let tx_id;
        let org = CONFIG.users[enrollId].org;
        // logger.info("enrollId:"+enrollId + " fnc:"+fnc + " args:"+ args);

        return new Promise(function(resolve, reject){
            helper.initObject(enrollId, org,false).then((clientObj)=>{
                client	 = clientObj.client;
                channel	 = clientObj.channel;
                targets = clientObj.targets;
                // ??? ????????? ???????????? ID??? ???????????? ????????? ??????????????? ?????????, ????????? ?????? ????????? ??????
                tx_id   = client.newTransactionID();

                //   return helper.getSubmitter(client,false, org, enrollId);
                return client.getUserContext(enrollId);
            }).then((submitter)=>{
                if(submitter){
                    let req = {
                        chaincodeId : CONFIG.chaincode.id,
                        txId : tx_id,
                        fcn: fnc,
                        args : args,
                        targets: targets
                    };
                    // ??? ??????????????? ????????? ??????
                    return channel.queryByChaincode(req);

                }else{
                    let emsg = "[blockchainService] query ERROR :" + fnc + " :enrollID :" + enrollId + " ***err :" + "User need enroll at first";
                    logger.error(emsg);
                    throw new Error(emsg);
                }

            }).then((payloads)=>{
                if(payloads){
                    // logger.info('Successfully query chaincode on the channel , payload : %s' ,payloads);
                    //we need only one result from payloads which are originated from several peers
                    // ??? ????????? ????????? ????????? ????????? ???????????? ????????? ??????
                    let result = Array.isArray(payloads) ? payloads[0] : payloads;
                    let buffer = new Buffer(result,'hex');
                    resolve({"result":buffer.toString('utf8')});


                }else{
                    let emsg = "[blockchainService] query ERROR :" + fnc +
                        " :enrollID :" + enrollId + " ***err :" + "response is null";
                    logger.error(emsg);
                    throw new Error(emsg);
                }
            }).catch((err)=>{
                err.message +=  "[" + fnc + "[args : " + args[0] + "] + [txid is : " + tx_id +  "] " ;
                reject(err);
            });
        });
    };



}

module.exports = BlockchainService;
