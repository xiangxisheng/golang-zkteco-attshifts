document.addEventListener('DOMContentLoaded',()=>{
  const form=document.getElementById('ym-form')
  const year=form?.querySelector('select[name="year"]')
  const month=form?.querySelector('select[name="month"]')
  ;[year,month].forEach(el=>el&&el.addEventListener('change',()=>form.submit()))
})

