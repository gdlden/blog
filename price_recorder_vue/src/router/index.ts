import {createRouter, createWebHistory} from 'vue-router'
import Article from "@/view/Article.vue";
import Api from "@/view/Api.vue";
import Login from "@/view/Login.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {path:"/",component: Api},
    {path:"/test",component:Article},
    {name: "login",
      path:"/login",
      component: Login}
  ],
})
let  isAuthenticated = false;
let userinfoStr = localStorage.getItem("user");
if(userinfoStr !== undefined && userinfoStr !== null) {
  const userInfoObj = JSON.parse(userinfoStr);
  if(userInfoObj.userId !== null) {
    isAuthenticated = true;
  }
}
router.beforeEach(async (to,from)=>{
  if(!isAuthenticated && to.path !== '/login') {
    console.log("未登录")
    return {name: 'login'}
  }else {
    console.log("已登录")
    return true
  }
})
router.onError((error, to, from)=>{
  console.error("error："+error)
})

export default router
