if (!Object.values) {
  Object.values = function (obj) {
    return Object.keys(obj).map(function (key) {
      return obj[key];
    })
  }
}


$(function () {
  $('.card .resource-info .name').on('click', function (e) {
    $(this).parents('.resource-info').toggleClass('expanded');
  });

  var expandMatch = window.location.search.match(/expand=(\w+)(\W|$)/);
  if (expandMatch && expandMatch[1] !== 'false' && expandMatch[1] !== '0') {
    $('.resource-info').addClass('expanded');
  }
});
