$(function () {

  // Check selected namespace options on page load
  const urlParams = new URLSearchParams(window.location.search);
  const currentNamespaces = urlParams.getAll('ns');
  currentNamespaces.forEach(ns => {
    $(`input#namespace-${ns}`).prop('checked', true);
  });

  // Handle new filter submissions
  $('#namespaceFiltersForm').on('submit', e => {
    e.preventDefault();
    let newParams = new URLSearchParams();
    $('#namespaceFiltersForm input[type="checkbox"]').each((index, checkbox) => {
      if (checkbox.checked) {
        newParams.append('ns', checkbox.name);
      }
    });
    window.location = new URL(`?${newParams.toString()}`, window.location).toString();
  });
});

