<script setup>
import {login, user} from "@/api/Login.js";
import {reactive, ref} from "vue";
import {useUserStore} from "@/stores/userStore.js";
import router from "@/router/index.js";

let useUserStore1 = useUserStore();
const cancel = ()=>{

}

const loginAction = ()=>{
  const userData = login(user);
  userData.then(res=>{
    console.log(res.token)
    console.log(res.user.userId)
    useUserStore1.userInfo = res;
    alert("登录成功！")
    router.push("/")
  })
}
</script>
<template>
<div class="flex items-center justify-center h-screen bg-transparent">
<div>
  <div class="flex">
    <div class="label-area ">
      <label class="font-sans">用户名：</label> <div></div>
      <label class="font-sans">密&emsp;码：</label><div></div>
<!--      <label class="font-sans">邮&emsp;箱：</label><div></div>-->

    </div>
    <div class="input-area">
      <input type="text" name="username" v-model="user.username" class="border-2 border-solid border-red-2"><div></div>
      <input type="password" name="password" v-model="user.password" class="border-2 border-solid border-red-2"><div></div>
<!--      <input type="text" name="email" v-model="user.email" class="border-2 border-solid border-red-2"><div></div>-->
    </div>
  </div>

  <div></div>


  <div class="flex items-center justify-center">
    <div><input type="button" value="重置" @click="cancel" class="animate-pulse"></div>
    &emsp;&emsp;
    <div><input type="button" value="登录" @click="loginAction" class="animate-pulse"></div>
  </div>
</div>
</div>
</template>
<style scoped>
</style>
