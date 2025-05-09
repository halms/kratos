/*
Ory Identities API

This is the API specification for Ory Identities with features such as registration, login, recovery, account verification, profile settings, password reset, identity management, session management, email and sms delivery, and more.

API version:
Contact: office@ory.sh
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package client

import (
	"encoding/json"
)

// checks if the IdentityWithCredentials type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &IdentityWithCredentials{}

// IdentityWithCredentials Create Identity and Import Credentials
type IdentityWithCredentials struct {
	Oidc                 *IdentityWithCredentialsOidc     `json:"oidc,omitempty"`
	Password             *IdentityWithCredentialsPassword `json:"password,omitempty"`
	Saml                 *IdentityWithCredentialsSaml     `json:"saml,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _IdentityWithCredentials IdentityWithCredentials

// NewIdentityWithCredentials instantiates a new IdentityWithCredentials object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewIdentityWithCredentials() *IdentityWithCredentials {
	this := IdentityWithCredentials{}
	return &this
}

// NewIdentityWithCredentialsWithDefaults instantiates a new IdentityWithCredentials object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewIdentityWithCredentialsWithDefaults() *IdentityWithCredentials {
	this := IdentityWithCredentials{}
	return &this
}

// GetOidc returns the Oidc field value if set, zero value otherwise.
func (o *IdentityWithCredentials) GetOidc() IdentityWithCredentialsOidc {
	if o == nil || IsNil(o.Oidc) {
		var ret IdentityWithCredentialsOidc
		return ret
	}
	return *o.Oidc
}

// GetOidcOk returns a tuple with the Oidc field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IdentityWithCredentials) GetOidcOk() (*IdentityWithCredentialsOidc, bool) {
	if o == nil || IsNil(o.Oidc) {
		return nil, false
	}
	return o.Oidc, true
}

// HasOidc returns a boolean if a field has been set.
func (o *IdentityWithCredentials) HasOidc() bool {
	if o != nil && !IsNil(o.Oidc) {
		return true
	}

	return false
}

// SetOidc gets a reference to the given IdentityWithCredentialsOidc and assigns it to the Oidc field.
func (o *IdentityWithCredentials) SetOidc(v IdentityWithCredentialsOidc) {
	o.Oidc = &v
}

// GetPassword returns the Password field value if set, zero value otherwise.
func (o *IdentityWithCredentials) GetPassword() IdentityWithCredentialsPassword {
	if o == nil || IsNil(o.Password) {
		var ret IdentityWithCredentialsPassword
		return ret
	}
	return *o.Password
}

// GetPasswordOk returns a tuple with the Password field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IdentityWithCredentials) GetPasswordOk() (*IdentityWithCredentialsPassword, bool) {
	if o == nil || IsNil(o.Password) {
		return nil, false
	}
	return o.Password, true
}

// HasPassword returns a boolean if a field has been set.
func (o *IdentityWithCredentials) HasPassword() bool {
	if o != nil && !IsNil(o.Password) {
		return true
	}

	return false
}

// SetPassword gets a reference to the given IdentityWithCredentialsPassword and assigns it to the Password field.
func (o *IdentityWithCredentials) SetPassword(v IdentityWithCredentialsPassword) {
	o.Password = &v
}

// GetSaml returns the Saml field value if set, zero value otherwise.
func (o *IdentityWithCredentials) GetSaml() IdentityWithCredentialsSaml {
	if o == nil || IsNil(o.Saml) {
		var ret IdentityWithCredentialsSaml
		return ret
	}
	return *o.Saml
}

// GetSamlOk returns a tuple with the Saml field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IdentityWithCredentials) GetSamlOk() (*IdentityWithCredentialsSaml, bool) {
	if o == nil || IsNil(o.Saml) {
		return nil, false
	}
	return o.Saml, true
}

// HasSaml returns a boolean if a field has been set.
func (o *IdentityWithCredentials) HasSaml() bool {
	if o != nil && !IsNil(o.Saml) {
		return true
	}

	return false
}

// SetSaml gets a reference to the given IdentityWithCredentialsSaml and assigns it to the Saml field.
func (o *IdentityWithCredentials) SetSaml(v IdentityWithCredentialsSaml) {
	o.Saml = &v
}

func (o IdentityWithCredentials) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o IdentityWithCredentials) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Oidc) {
		toSerialize["oidc"] = o.Oidc
	}
	if !IsNil(o.Password) {
		toSerialize["password"] = o.Password
	}
	if !IsNil(o.Saml) {
		toSerialize["saml"] = o.Saml
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *IdentityWithCredentials) UnmarshalJSON(data []byte) (err error) {
	varIdentityWithCredentials := _IdentityWithCredentials{}

	err = json.Unmarshal(data, &varIdentityWithCredentials)

	if err != nil {
		return err
	}

	*o = IdentityWithCredentials(varIdentityWithCredentials)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "oidc")
		delete(additionalProperties, "password")
		delete(additionalProperties, "saml")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableIdentityWithCredentials struct {
	value *IdentityWithCredentials
	isSet bool
}

func (v NullableIdentityWithCredentials) Get() *IdentityWithCredentials {
	return v.value
}

func (v *NullableIdentityWithCredentials) Set(val *IdentityWithCredentials) {
	v.value = val
	v.isSet = true
}

func (v NullableIdentityWithCredentials) IsSet() bool {
	return v.isSet
}

func (v *NullableIdentityWithCredentials) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIdentityWithCredentials(val *IdentityWithCredentials) *NullableIdentityWithCredentials {
	return &NullableIdentityWithCredentials{value: val, isSet: true}
}

func (v NullableIdentityWithCredentials) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIdentityWithCredentials) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
