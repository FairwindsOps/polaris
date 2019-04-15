if (!Object.values) {
  Object.values = function (obj) {
    return Object.keys(obj).map(function (key) {
      return obj[key];
    })
  }
}

$(function () {
  var namespaceChart = new Chart("namespaceScoreChart", {
    type: 'bar',
    data: {
      labels: Object.keys(fairwindsAuditData.NamespacedResults),
      datasets: [{
        label: 'Passing',
        data: Object.values(fairwindsAuditData.NamespacedResults)
          .map(function (r) { return r.Summary.Successes }),
        backgroundColor: '#8BD2DC',
      }, {
        label: 'Warning',
        data: Object.values(fairwindsAuditData.NamespacedResults)
          .map(function (r) { return r.Summary.Warnings }),
        backgroundColor: '#f26c21',
      }, {
        label: 'Failing',
        data: Object.values(fairwindsAuditData.NamespacedResults)
          .map(function (r) { return r.Summary.Errors }),
        backgroundColor: '#a11f4c',
      }]
    },
    options: {
      legend: {
        display: false,
      },
      scales: {
        xAxes: [{
          stacked: true,
        }],
        yAxes: [{
          stacked: true,
          ticks: {
            beginAtZero: true
          }
        }]
      }
    }
  });

  var score = fairwindsAuditData.ClusterSummary.Successes / (
    fairwindsAuditData.ClusterSummary.Successes +
    fairwindsAuditData.ClusterSummary.Warnings +
    fairwindsAuditData.ClusterSummary.Errors);
  score = Math.round(score * 100);

  var clusterChart = new Chart("clusterScoreChart", {
    type: 'doughnut',
    data: {
      labels: ["Passing", "Warning", "Error"],
      datasets: [{
        data: [
          fairwindsAuditData.ClusterSummary.Successes,
          fairwindsAuditData.ClusterSummary.Warnings,
          fairwindsAuditData.ClusterSummary.Errors,
        ],
        backgroundColor: ['#8BD2DC', '#f26c21', '#a11f4c'],
      }]
    },
    options: {
      // responsive: false,
      cutoutPercentage: 75,
      legend: {
        display: false,
      },
      elements: {
        center: {
          text: score,
          color: '#333', //Default black
          fontStyle: 'Muli', //Default Arial
          sidePadding: 40, //Default 20 (as a percentage)
        }
      }
    }
  });
});

