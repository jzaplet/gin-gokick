package events

//tsgen:assets/shared/Tracking/types/TrackedEvent.ts TrackedEvent union
type TrackedEvent string

const TrackedEventSignUp TrackedEvent = "sign_up"

const TrackedEventLogin TrackedEvent = "login"

//tsgen:assets/shared/Tracking/types/AdsConversion.ts AdsConversion union
type AdsConversion string

const AdsConversionSignUp AdsConversion = "sign_up"

var AdsConversions = []AdsConversion{AdsConversionSignUp}
