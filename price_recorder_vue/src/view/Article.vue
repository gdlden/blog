<script setup>
import {onMounted, ref} from "vue";
import {getAllArticle} from "@/api/Article.js";
const tableData = ref([])
onMounted(()=>{
  getAllArticle().then(res=>{
    console.log(res)
    tableData.value = res.data
  })
})
</script>
<template>
  <table>
    <tr v-for="(article,index) in tableData">
      <td>{{article.title}}</td>
      <td>{{article.name}}</td>
    </tr>
  </table>
</template>
