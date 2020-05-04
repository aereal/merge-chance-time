import firebase from "firebase/app"
import "firebase/auth"

const options = {
  apiKey: "AIzaSyBEAHMliUHu1oURysWwI0q2nL0hSugBOjQ",
  authDomain: "merge-chance-time.firebaseapp.com",
  projectId: "merge-chance-time",
}

firebase.initializeApp(options)

export const auth = (): firebase.auth.Auth => {
  if (firebase.apps.length === 0) {
    firebase.initializeApp(options)
  }
  return firebase.auth()
}

export const ghProvider = new firebase.auth.GithubAuthProvider()

export const signInWithFirebaseAuth = async (): Promise<void> => {
  await auth().signInWithRedirect(ghProvider)
}
