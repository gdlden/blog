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
  <div class="max-w-[980px] mx-auto px-5 py-12">
    <h2 style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 40px; font-weight: 600; line-height: 1.10; color: #1d1d1f; margin-bottom: 24px;">
      文章列表
    </h2>
    <div class="bg-white" style="border-radius: 8px; box-shadow: rgba(0, 0, 0, 0.22) 3px 5px 30px 0px; overflow: hidden;">
      <table class="w-full">
        <thead>
          <tr style="border-bottom: 1px solid rgba(0, 0, 0, 0.04);">
            <th class="px-5 py-3 text-left" style="font-size: 14px; font-weight: 600; letter-spacing: -0.224px; color: rgba(0, 0, 0, 0.48);">标题</th>
            <th class="px-5 py-3 text-left" style="font-size: 14px; font-weight: 600; letter-spacing: -0.224px; color: rgba(0, 0, 0, 0.48);">名称</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(article,index) in tableData" :key="index" style="border-bottom: 1px solid rgba(0, 0, 0, 0.04);">
            <td class="px-5 py-3" style="font-size: 14px; letter-spacing: -0.224px; color: #1d1d1f;">{{article.title}}</td>
            <td class="px-5 py-3" style="font-size: 14px; letter-spacing: -0.224px; color: rgba(0, 0, 0, 0.8);">{{article.name}}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
