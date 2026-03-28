<script setup>
import {onMounted, ref} from "vue";

const count = ref(0)
function add() {
  count.value++
}
onMounted(()=>{
  console.log("vue已经挂载完毕")
})
</script>
<template>
  <button @click="add">count: {{count}}</button>
  <div class="w-4/5 font-bold font-sans">测试一下</div>
  <div>测试一下</div>
</template>
