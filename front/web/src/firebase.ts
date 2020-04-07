import firebase from "firebase/app"
import "firebase/auth"

firebase.initializeApp({
  apiKey: "AIzaSyBEAHMliUHu1oURysWwI0q2nL0hSugBOjQ",
  authDomain: "merge-chance-time.firebaseapp.com",
  projectId: "merge-chance-time",
})

export const auth = (): firebase.auth.Auth => {
  if (firebase.apps.length === 0) {
    firebase.initializeApp({
      apiKey: "AIzaSyBEAHMliUHu1oURysWwI0q2nL0hSugBOjQ",
      authDomain: "merge-chance-time.firebaseapp.com",
      projectId: "merge-chance-time",
    })
  }

  return firebase.auth()
}
