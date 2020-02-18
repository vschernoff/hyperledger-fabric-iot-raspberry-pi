package ca

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hlf-iot/config"
	"regexp"
	"sync"
)

//response /ca/tbs-csr
var responseTbsCsrRegex = regexp.MustCompile(`{"tbs_csr_bytes"([\s\S]*?)"}|{"tbs_csr_hash"([\s\S]*?)"}`)

//response /ca/enroll-csr
var responseEnrollCsrRegex = regexp.MustCompile(`{"user_cert"([\s\S]*?)"}`)

//response /tx/proposal
var responseProposalRegex = regexp.MustCompile(`{"proposal_bytes"([\s\S]*?)"}|{"proposal_hash"([\s\S]*?)"}`)

//response /tx/prepare-broadcast
var responseBroadcastPayloadRegex = regexp.MustCompile(`{"payload_bytes"([\s\S]*?)"}|{"payload_hash"([\s\S]*?)"}`)

type caCreds struct {
	Login    string
	Password string
}

type TbsCsrReq struct {
	X     string `json:"x"`
	Y     string `json:"y"`
	Login string `json:"login"`
}

type EnrollCsrReq struct {
	Login       string `json:"login"`
	Password    string `json:"password"`
	TbsCsrBytes string `json:"tbs_csr_bytes"`
	R           string `json:"r"`
	S           string `json:"s"`
}

type ProposalReq struct {
	ChannelId   string   `json:"channel_id"`
	ChaincodeId string   `json:"chaincode_id"`
	MspId       string   `json:"msp_id"`
	Fcn         string   `json:"fcn"`
	Args        []string `json:"args"`
	UserCert    string   `json:"user_cert"`
}

type BroadcastPayloadReq struct {
	ProposalBytes string   `json:"proposal_bytes"`
	Peers         []string `json:"peers"`
	R             string   `json:"r"`
	S             string   `json:"s"`
}

type BroadcastReq struct {
	PayloadBytes string `json:"payload_bytes"`
	R            string `json:"r"`
	S            string `json:"s"`
}

type TbsCsrResponse struct {
	TbsCsrBytes string `json:"tbs_csr_bytes"`
	TbsCsrHash  string `json:"tbs_csr_hash"`
}

type Signature struct {
	R string `json:"r"`
	S string `json:"s"`
}

type UserCertificate struct {
	Certificate string `json:"user_cert"`
}

type Proposal struct {
	ProposalBytes string `json:"proposal_bytes"`
	ProposalHash  string `json:"proposal_hash"`
}

type BroadcastPayload struct {
	BroadcastPayloadBytes string `json:"payload_bytes"`
	BroadcastPayloadHash  string `json:"payload_hash"`
}

type Ca struct {
	PrivateKey                *ecdsa.PrivateKey `json:"privatekey"`
	CaCreds                   caCreds           `json:"cacreds"`
	TbsCsrResponse            *TbsCsrResponse   `json:"tbscsrresponse"`
	TbsCsrSignature           Signature         `json:"tbscsrsignature"`
	ProposalSignature         Signature         `json:"proposalsignature"`
	BroadcastPayloadSignature Signature         `json:"broadcastpayloadsignature"`
	UserCertificate           *UserCertificate  `json:"usercertificate"`
	Proposal                  *Proposal         `json:"proposal"`
	BroadcastPayload          *BroadcastPayload `json:"broadcastpayload"`
}

var instance *Ca
var once sync.Once

func GetInstance() *Ca {
	once.Do(func() {
		instance = &Ca{}
		instance.CaCreds.Login = config.CA_LOGIN
		instance.CaCreds.Password = config.CA_PASSWORD
	})
	return instance
}

// Private key generation
func (ca *Ca) GeneratePrivateKey() {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("ecdsa failed to generate private key")
	}

	ca.PrivateKey = privateKey
}

// Creates TBS CSR from public key and user data
func (ca *Ca) TbsCsr() *config.GprsSendElementStructure {
	tbsCsrReqJson, _ := json.Marshal(&TbsCsrReq{
		X:     ca.PrivateKey.X.Text(16),
		Y:     ca.PrivateKey.Y.Text(16),
		Login: ca.CaCreds.Login,
	})

	return &config.GprsSendElementStructure{
		Data:       string(tbsCsrReqJson),
		Url:        config.CA_TBS_CSR_URL,
		HttpAction: config.GPRS_HTTPACTION_POST,
		CheckFcn:   ca.CheckRequestTbsCsrToBC,
	}
}

// Sign TBS CSR
func (ca *Ca) SignTbsCsr() {
	hash, _ := hex.DecodeString(ca.TbsCsrResponse.TbsCsrHash)
	r, s, err := ecdsa.Sign(rand.Reader, ca.PrivateKey, hash)
	if err != nil {
		fmt.Println(err)
	}

	ca.TbsCsrSignature.R = r.Text(16)
	ca.TbsCsrSignature.S = s.Text(16)
}

// Enroll with signed CSR
func (ca *Ca) EnrollCsr() *config.GprsSendElementStructure {
	enrollCsrReqJson, _ := json.Marshal(&EnrollCsrReq{
		Login:       ca.CaCreds.Login,
		Password:    ca.CaCreds.Password,
		TbsCsrBytes: ca.TbsCsrResponse.TbsCsrBytes,
		R:           ca.TbsCsrSignature.R,
		S:           ca.TbsCsrSignature.S,
	})
	enrollCsrReqJsonStr := string(enrollCsrReqJson)

	return &config.GprsSendElementStructure{
		Data:       enrollCsrReqJsonStr,
		Url:        config.CA_ENROLL_CSR_URL,
		HttpAction: config.GPRS_HTTPACTION_POST,
		CheckFcn:   ca.CheckRequestEnrollCsrToBC,
	}
}

func (ca *Ca) ProposalReq(fcn string, args []string) *config.GprsSendElementStructure {
	proposalReqJson, _ := json.Marshal(&ProposalReq{
		Fcn:         fcn,
		Args:        args,
		ChaincodeId: config.CHAINCODE_ID,
		ChannelId:   config.CHANNEL_ID,
		MspId:       config.MSP_ID,
		UserCert:    ca.UserCertificate.Certificate,
	})
	proposalReqJsonStr := string(proposalReqJson)

	return &config.GprsSendElementStructure{
		Data:       proposalReqJsonStr,
		Url:        config.CA_PROPOSAL_URL,
		HttpAction: config.GPRS_HTTPACTION_POST,
		CheckFcn:   ca.CheckRequestProposalToBC,
	}
}

func (ca *Ca) SignProposal() {
	hash, _ := hex.DecodeString(ca.Proposal.ProposalHash)
	r, s, err := ecdsa.Sign(rand.Reader, ca.PrivateKey, hash)
	if err != nil {
		fmt.Println(err)
	}

	ca.ProposalSignature.R = r.Text(16)
	ca.ProposalSignature.S = s.Text(16)
}

func (ca *Ca) BroadcastPayloadReq() *config.GprsSendElementStructure {
	broadcastPayloadReqJson, _ := json.Marshal(&BroadcastPayloadReq{
		ProposalBytes: ca.Proposal.ProposalBytes,
		Peers:         config.EndorsementPeers,
		R:             ca.ProposalSignature.R,
		S:             ca.ProposalSignature.S,
	})
	broadcastPayloadReqJsonStr := string(broadcastPayloadReqJson)

	return &config.GprsSendElementStructure{
		Data:       broadcastPayloadReqJsonStr,
		Url:        config.CA_BROADCAST_PAYLOAD_URL,
		HttpAction: config.GPRS_HTTPACTION_POST,
		CheckFcn:   ca.CheckRequestBroadcastPayloadToBC,
	}
}

func (ca *Ca) SignBroadcastPayload() {
	hash, _ := hex.DecodeString(ca.BroadcastPayload.BroadcastPayloadHash)
	r, s, err := ecdsa.Sign(rand.Reader, ca.PrivateKey, hash)
	if err != nil {
		fmt.Println(err)
	}

	ca.BroadcastPayloadSignature.R = r.Text(16)
	ca.BroadcastPayloadSignature.S = s.Text(16)
}

func (ca *Ca) BroadcastReq() *config.GprsSendElementStructure {
	broadcastReqJson, _ := json.Marshal(&BroadcastReq{
		PayloadBytes: ca.BroadcastPayload.BroadcastPayloadBytes,
		R:            ca.BroadcastPayloadSignature.R,
		S:            ca.BroadcastPayloadSignature.S,
	})
	broadcastReqJsonStr := string(broadcastReqJson)

	return &config.GprsSendElementStructure{
		Data:       broadcastReqJsonStr,
		Url:        config.CA_BROADCAST_URL,
		HttpAction: config.GPRS_HTTPACTION_POST,
		CheckFcn:   config.CheckInsertToBC,
	}
}

func (ca *Ca) CheckRequestTbsCsrToBC(buffer string) bool {
	fmt.Println("output")
	fmt.Println(buffer)
	response := responseTbsCsrRegex.FindAllStringSubmatch(buffer, -1)
	if len(response) > 0 {
		tbsCsrResponse := &TbsCsrResponse{}
		if err := json.Unmarshal([]byte(response[0][0]), &tbsCsrResponse); err != nil {
			fmt.Printf("unable to unmarshal CheckRequestTbsCsrToBC result: %s", err.Error())
		}
		ca.TbsCsrResponse = tbsCsrResponse

		return true
	}

	return false
}

func (ca *Ca) CheckRequestEnrollCsrToBC(buffer string) bool {
	fmt.Println("CheckRequestEnrollCsrToBC output")
	fmt.Println(buffer)
	response := responseEnrollCsrRegex.FindAllStringSubmatch(buffer, -1)
	if len(response) > 0 {
		userCertificate := &UserCertificate{}
		if err := json.Unmarshal([]byte(response[0][0]), &userCertificate); err != nil {
			fmt.Printf("unable to unmarshal CheckRequestEnrollCsrToBC result: %s", err.Error())
		}
		userCertificateByte, _ := config.B64Decode(userCertificate.Certificate)
		userCertificate.Certificate = string(userCertificateByte)
		ca.UserCertificate = userCertificate

		return true
	}

	return false
}

func (ca *Ca) CheckRequestProposalToBC(buffer string) bool {
	fmt.Println("CheckRequestProposalToBC output")
	fmt.Println(buffer)
	response := responseProposalRegex.FindAllStringSubmatch(buffer, -1)
	if len(response) > 0 {
		proposal := &Proposal{}
		if err := json.Unmarshal([]byte(response[0][0]), &proposal); err != nil {
			fmt.Printf("unable to unmarshal CheckRequestProposalToBC result: %s", err.Error())
		}
		ca.Proposal = proposal

		return true
	}

	return false
}

func (ca *Ca) CheckRequestBroadcastPayloadToBC(buffer string) bool {
	fmt.Println("CheckRequestBroadcastPayloadToBC output")
	fmt.Println(buffer)
	response := responseBroadcastPayloadRegex.FindAllStringSubmatch(buffer, -1)
	if len(response) > 0 {
		broadcastPayload := &BroadcastPayload{}
		if err := json.Unmarshal([]byte(response[0][0]), &broadcastPayload); err != nil {
			fmt.Printf("unable to unmarshal CheckRequestBroadcastPayloadToBC result: %s", err.Error())
		}
		ca.BroadcastPayload = broadcastPayload

		return true
	}

	return false
}
