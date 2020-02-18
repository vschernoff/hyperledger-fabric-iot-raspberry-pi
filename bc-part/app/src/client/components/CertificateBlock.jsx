import React from 'react';
import EC from 'elliptic';

let ec = new EC.ec('p256');

class DisplayData extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            tbsCsrHash: "",
            tbsCsrBytes: "",

            tbsCsrSignature: {
                r: "",
                s: ""
            },

            caCreds: {
                login: "admin",
                password: "adminpw",
                email: ""
            },

            enrollSuccess: false,

            key: "",
            privateKeyHex: "",
            userCert: ``,
            proposalB64: "",
            proposalHash: "",
            proposalSignature: {
                r: "",
                s: ""
            },

            proposalRequest: {
                channel_id: "common",
                chaincode_id: "hlf_iot_cc",
                msp_id: "hlfiotMSP",
                fcn: "addIotCertificate",
                args: ""
            },

            endorsementPeers: "hlfiot/peer0,device/peer0",

            broadcastPayloadB64: "",
            broadcastPayloadHash: "",
            broadcastPayloadSignature: {
                r: "",
                s: ""
            },

            broadcastResponse: ""
        };

        this.generateCertificate = this.generateCertificate.bind(this);
        this.uploadCertificate = this.uploadCertificate.bind(this);
    }

    generateCertificate() {
        this.generatePrivateKey();
        this.tbsCsr();
    }

    uploadCertificate() {
        this.proposal();
    }

    publicKeyX() {
        return this.state.privateKeyHex.length
            ? ec
                .keyFromPrivate(this.state.privateKeyHex, "hex")
                .getPublic()
                .getX()
                .toString(16)
            : "";
    }

    publicKeyY() {
        return this.state.privateKeyHex.length
            ? ec
                .keyFromPrivate(this.state.privateKeyHex, "hex")
                .getPublic()
                .getY()
                .toString(16)
            : "";
    }

    tbsCsr() {
        let req = {
            X: this.publicKeyX(),
            Y: this.publicKeyY(),
            Login: this.state.caCreds.login,
            Email: Date.now().toString()
        };

        fetch("/api/ca/tbs-csr", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(req)
        })
            .then(response => response.json())
            .then(responseJson => {
                this.state.tbsCsrBytes = atob(responseJson.tbs_csr_bytes);
                this.state.tbsCsrHash = responseJson.tbs_csr_hash;

                this.signTbsCsr();
                this.enrollCsr();
            });
    }

    signTbsCsr() {
        let keyPriv = ec.keyFromPrivate(this.state.privateKeyHex, "hex");
        let signature = keyPriv.sign(this.state.tbsCsrHash);

        this.state.tbsCsrSignature.r = signature.r.toString(16);
        this.state.tbsCsrSignature.s = signature.s.toString(16);
    }

    enrollCsr() {
        let req = {
            login: this.state.caCreds.login,
            password: this.state.caCreds.password,
            tbs_csr_bytes: btoa(this.state.tbsCsrBytes),
            r: this.state.tbsCsrSignature.r,
            s: this.state.tbsCsrSignature.s
        };

        fetch("/api/ca/enroll-csr", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(req)
        })
            .then(response => response.json())
            .then(responseJson => {
                let pem = atob(responseJson.user_cert);
                if (pem.includes("BEGIN CERTIFICATE")) {
                    this.setState({
                        userCert: pem
                    });
                    this.state.enrollSuccess = true;
                    setTimeout(() => {
                        this.state.enrollSuccess = false;
                    }, 3000);
                }
            });
    }

    generatePrivateKey() {
        let key = ec.genKeyPair();
        this.state.privateKeyHex = key.getPrivate().toString(16);
    }

    proposal() {
        let req = {...this.state.proposalRequest};
        req.user_cert = req.args = this.state.userCert;
        req.args = req.args.split(",");

        fetch("/api/tx/proposal", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(req)
        })
            .then(response => response.json())
            .then(responseJson => {
                this.state.proposalB64 = responseJson.proposal_bytes;
                this.state.proposalHash = responseJson.proposal_hash;

                let keyPriv = ec.keyFromPrivate(this.state.privateKeyHex, "hex");
                let signature = keyPriv.sign(this.state.proposalHash);

                this.state.proposalSignature.r = signature.r.toString(16);
                this.state.proposalSignature.s = signature.s.toString(16);

                this.broadcastPayload();
            });
    }

    broadcastPayload() {
        let req = {
            proposal_bytes: this.state.proposalB64,
            peers: this.state.endorsementPeers.split(","),
            ...this.state.proposalSignature
        };

        fetch("/api/tx/prepare-broadcast", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(req)
        })
            .then(response => response.json())
            .then(responseJson => {
                this.state.broadcastPayloadB64 = responseJson.payload_bytes;
                this.state.broadcastPayloadHash = responseJson.payload_hash;

                let keyPriv = ec.keyFromPrivate(this.state.privateKeyHex, "hex");
                let signature = keyPriv.sign(this.state.broadcastPayloadHash);

                this.state.broadcastPayloadSignature.r = signature.r.toString(16);
                this.state.broadcastPayloadSignature.s = signature.s.toString(16);

                this.broadcast();
            });
    }

    broadcast() {
        let req = {
            payload_bytes: this.state.broadcastPayloadB64,
            ...this.state.broadcastPayloadSignature
        };

        fetch("/api/tx/broadcast", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(req)
        })
            .then(response => response.json())
            .then(responseJson => {
                this.setState({broadcastResponse: JSON.stringify(responseJson)});
            });
    }

    render() {

        const {privateKeyHex, userCert, broadcastResponse} = this.state;

        let pemData = 'data:application/x-pem-file;charset=utf-8,' + encodeURIComponent(userCert);
        let certificateFileName = privateKeyHex + ".pem";

        return (
            <div>
                <div>
                    <div className='form-group'>
                        <div className='input-group justify-content-center '>
                            <button className="btn btn-sm btn-primary w-50 m-2" title="Generate Certificate"
                                    onClick={() => this.generateCertificate()}>
                                Generate Certificate
                            </button>
                        </div>
                    </div>
                    {(userCert) && (
                        <div className='form-group'>
                            <div className="input-group justify-content-center">
                <textarea name="certificate" id="certificate" cols="60" rows="10">
                  {userCert}
                </textarea>
                            </div>
                        </div>
                    )}
                    {(userCert) && (
                        <div className='form-group'>
                            <div className='input-group justify-content-center '>
                                <button className="btn btn-sm btn-primary w-50 m-2" title="Upload Certificate to BC"
                                        onClick={() => this.uploadCertificate()}>
                                    Upload Certificate to BC
                                </button>
                            </div>
                        </div>
                    )}
                    {(broadcastResponse) && (
                        <div className='form-group'>
                            <div className="input-group justify-content-center">
                <textarea name="broadcastResponse" id="broadcastResponse" cols="60" rows="10">
                  {broadcastResponse}
                </textarea>
                            </div>
                        </div>
                    )}
                    {(broadcastResponse) && (
                        <div className='form-group'>
                            <div className="input-group justify-content-center">
                                <a className="btn btn-sm btn-primary w-50 m-2" title="Download certificate"
                                   href={pemData} target="_blank" download={certificateFileName}>
                                    Download certificate
                                </a>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        );
    }
}

export {DisplayData as CertificateBlock};
