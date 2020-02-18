package ca

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hlf-iot/config"
	"hlf-iot/helpers/fswrapper"
	"hlf-iot/helpers/httpwrapper"
	"net/url"
	"sync"
)

type caCreds struct {
	Login    string
	Password string
	Email    string
}

type TbsCsrReq struct {
	X     string `json:"x"`
	Y     string `json:"y"`
	Login string `json:"login"`
	Email string `json:"email"`
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
		instance.CaCreds.Email = config.GetCustomField(config.CA_CUSTOM_FIELD)
	})
	return instance
}

// Private key generation
func (ca *Ca) GeneratePrivateKey() error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	ca.PrivateKey = privateKey

	return nil
}

// Creates TBS CSR from public key and user data
func (ca *Ca) TbsCsr() (*httpwrapper.SendElementStructure, error) {
	if ca.PrivateKey.X == nil || len(ca.PrivateKey.X.Text(16)) == 0 {
		return nil, errors.New("privateKey error")
	}

	if len(ca.CaCreds.Login) == 0 || len(ca.CaCreds.Email) == 0 {
		return nil, errors.New("ca credentials are not initialized")
	}

	tbsCsrReqJson, _ := json.Marshal(&TbsCsrReq{
		X:     ca.PrivateKey.X.Text(16),
		Y:     ca.PrivateKey.Y.Text(16),
		Login: ca.CaCreds.Login,
		Email: ca.CaCreds.Email,
	})

	return &httpwrapper.SendElementStructure{
		Data:     string(tbsCsrReqJson),
		Url:      config.CA_TBS_CSR_URL,
		CheckFcn: ca.CheckRequestTbsCsrToBC,
	}, nil
}

// Sign TBS CSR
func (ca *Ca) SignTbsCsr() error {
	hash, _ := hex.DecodeString(ca.TbsCsrResponse.TbsCsrHash)
	r, s, err := ecdsa.Sign(rand.Reader, ca.PrivateKey, hash)
	if err != nil {
		fmt.Println(err)
	}

	ca.TbsCsrSignature.R = r.Text(16)
	ca.TbsCsrSignature.S = s.Text(16)

	if len(ca.TbsCsrSignature.R) == 0 || len(ca.TbsCsrSignature.S) == 0 {
		return errors.New("tbsCsrSignature error")
	}

	return nil
}

// Enroll with signed CSR
func (ca *Ca) EnrollCsr() (*httpwrapper.SendElementStructure, error) {
	enrollCsrReqJson, err := json.Marshal(&EnrollCsrReq{
		Login:       ca.CaCreds.Login,
		Password:    ca.CaCreds.Password,
		TbsCsrBytes: ca.TbsCsrResponse.TbsCsrBytes,
		R:           ca.TbsCsrSignature.R,
		S:           ca.TbsCsrSignature.S,
	})
	if err != nil {
		return nil, err
	}

	enrollCsrReqJsonStr := string(enrollCsrReqJson)

	return &httpwrapper.SendElementStructure{
		Data:     enrollCsrReqJsonStr,
		Url:      config.CA_ENROLL_CSR_URL,
		CheckFcn: ca.CheckRequestEnrollCsrToBC,
	}, nil
}

func (ca *Ca) ProposalReq(fcn string, args []string) (*httpwrapper.SendElementStructure, error) {
	proposalReqJson, err := json.Marshal(&ProposalReq{
		Fcn:         fcn,
		Args:        args,
		ChaincodeId: config.CHAINCODE_ID,
		ChannelId:   config.CHANNEL_ID,
		MspId:       config.MSP_ID,
		UserCert:    ca.UserCertificate.Certificate,
	})
	if err != nil {
		return nil, err
	}

	proposalReqJsonStr := string(proposalReqJson)

	return &httpwrapper.SendElementStructure{
		Data:     proposalReqJsonStr,
		Url:      config.CA_PROPOSAL_URL,
		CheckFcn: ca.CheckRequestProposalToBC,
	}, nil
}

func (ca *Ca) SignProposal() error {
	hash, err := hex.DecodeString(ca.Proposal.ProposalHash)
	if err != nil {
		return err
	}

	r, s, err := ecdsa.Sign(rand.Reader, ca.PrivateKey, hash)
	if err != nil {
		return err
	}

	ca.ProposalSignature.R = r.Text(16)
	ca.ProposalSignature.S = s.Text(16)

	return nil
}

func (ca *Ca) BroadcastPayloadReq() (*httpwrapper.SendElementStructure, error) {
	broadcastPayloadReqJson, err := json.Marshal(&BroadcastPayloadReq{
		ProposalBytes: ca.Proposal.ProposalBytes,
		Peers:         config.EndorsementPeers,
		R:             ca.ProposalSignature.R,
		S:             ca.ProposalSignature.S,
	})
	if err != nil {
		return nil, err
	}

	broadcastPayloadReqJsonStr := string(broadcastPayloadReqJson)

	return &httpwrapper.SendElementStructure{
		Data:     broadcastPayloadReqJsonStr,
		Url:      config.CA_BROADCAST_PAYLOAD_URL,
		CheckFcn: ca.CheckRequestBroadcastPayloadToBC,
	}, nil
}

func (ca *Ca) SignBroadcastPayload() error {
	hash, err := hex.DecodeString(ca.BroadcastPayload.BroadcastPayloadHash)
	if err != nil {
		return err
	}

	r, s, err := ecdsa.Sign(rand.Reader, ca.PrivateKey, hash)
	if err != nil {
		return err
	}

	ca.BroadcastPayloadSignature.R = r.Text(16)
	ca.BroadcastPayloadSignature.S = s.Text(16)

	return nil
}

func (ca *Ca) BroadcastReq() (*httpwrapper.SendElementStructure, error) {
	broadcastReqJson, err := json.Marshal(&BroadcastReq{
		PayloadBytes: ca.BroadcastPayload.BroadcastPayloadBytes,
		R:            ca.BroadcastPayloadSignature.R,
		S:            ca.BroadcastPayloadSignature.S,
	})
	if err != nil {
		return nil, err
	}

	broadcastReqJsonStr := string(broadcastReqJson)

	return &httpwrapper.SendElementStructure{
		Data:     broadcastReqJsonStr,
		Url:      config.CA_BROADCAST_URL,
		CheckFcn: config.CheckInsertToBC,
	}, nil
}

func (ca *Ca) CheckRequestTbsCsrToBC(buffer string) (bool, error) {
	fmt.Println("CheckRequestTbsCsrToBC output")
	fmt.Println(buffer)

	var check bool
	tbsCsrResponse := &TbsCsrResponse{}
	if err := json.Unmarshal([]byte(buffer), &tbsCsrResponse); err != nil {
		return check, err
	}

	ca.TbsCsrResponse = tbsCsrResponse
	if len(ca.TbsCsrResponse.TbsCsrHash) > 0 {
		check = true
	} else {
		check = false
	}

	return check, nil
}

func (ca *Ca) CheckRequestEnrollCsrToBC(buffer string) (bool, error) {
	fmt.Println("CheckRequestEnrollCsrToBC output")
	fmt.Println(buffer)

	var check bool
	userCertificate := &UserCertificate{}
	if err := json.Unmarshal([]byte(buffer), &userCertificate); err != nil {
		return check, err
	}

	userCertificateByte, err := config.B64Decode(userCertificate.Certificate)
	if err != nil {
		return check, err
	}

	userCertificate.Certificate = string(userCertificateByte)
	ca.UserCertificate = userCertificate

	if len(ca.UserCertificate.Certificate) > 0 {
		check = true
	} else {
		check = false
	}

	return check, nil
}

func (ca *Ca) CheckRequestProposalToBC(buffer string) (bool, error) {
	var check bool
	proposal := &Proposal{}
	if err := json.Unmarshal([]byte(buffer), &proposal); err != nil {
		return check, err
	}

	ca.Proposal = proposal
	if len(ca.Proposal.ProposalHash) > 0 {
		check = true
	} else {
		check = false
	}

	return check, nil
}

func (ca *Ca) CheckRequestBroadcastPayloadToBC(buffer string) (bool, error) {
	var check bool
	broadcastPayload := &BroadcastPayload{}
	if err := json.Unmarshal([]byte(buffer), &broadcastPayload); err != nil {
		return check, err
	}

	ca.BroadcastPayload = broadcastPayload
	if len(ca.BroadcastPayload.BroadcastPayloadHash) > 0 {
		check = true
	} else {
		check = false
	}

	return check, nil
}

func (ca *Ca) GetCertificateFromKeyStorage() (bool, error) {
	var check bool

	keys, err := fswrapper.GetFilesDataFromKeyStorage(config.CERTIFICATE_FILE_EXTENSION)
	if err != nil {
		return check, err
	}

	for _, key := range keys {
		responseJson, err := httpwrapper.GetReq(config.API_BASE_URL + "channels/" + config.CHANNEL_ID + "/chaincodes/" + config.CHAINCODE_ID + "?fcn=" + config.FCN_NAME_CHECK_IOT_CERTIFICATE + "&peer=" + config.EndorsementPeers[0] + "&args=" + url.QueryEscape(key.Certificate))
		if err != nil {
			return check, err
		}

		var result map[string]interface{}
		json.Unmarshal([]byte(responseJson), &result)
		response := result["result"].(float64)

		if response == 1 {
			userCertificate := &UserCertificate{}
			userCertificate.Certificate = key.Certificate
			ca.UserCertificate = userCertificate

			privateKey, err := config.HexToPrivateKey(key.PrivateKeyHash)
			if err != nil {
				return check, err
			}

			fmt.Println("Certificate")
			fmt.Println(key.Certificate)
			fmt.Println("PrivateKeyHash")
			fmt.Println(key.PrivateKeyHash)

			ca.PrivateKey = privateKey
			check = true

			break
		}
	}

	return check, nil
}
