document.addEventListener("DOMContentLoaded", function(){
  setTimeout(function() {
    var link = document.getElementsByClassName('home-link')[0];
    linkClone = link.cloneNode(true);
    linkClone.href = "https://fairwinds.com";
    link.setAttribute('target', '_blank');
    console.log('set attr');
    link.parentNode.replaceChild(linkClone, link);
  }, 1000);
});

