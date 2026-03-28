import {ref, computed, watch} from 'vue'
import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', () => {
  const userInfo = ref("")
  watch(userInfo, (state) => {
    if(state !== null) {
      localStorage.setItem("user",JSON.stringify(state))
    }else {
      console.log(state)
    }
  },{deep: true})
  return {  userInfo }
})
