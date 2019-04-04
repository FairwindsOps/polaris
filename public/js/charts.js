$(function () {
  var namespaceChart = new Chart("namespaceScoreChart", {
    type: 'bar',
    data: {
      labels: Object.keys(AUDIT.NamespacedResults),
      datasets: [{
        label: 'Passing',
        data: Object.values(AUDIT.NamespacedResults).map(r => r.Summary.Successes),
        backgroundColor: '#8BD2DC',
      },{
        label: 'Warning',
        data: Object.values(AUDIT.NamespacedResults).map(r => r.Summary.Warnings),
        backgroundColor: '#f26c21',
      },{
        label: 'Failing',
        data: Object.values(AUDIT.NamespacedResults).map(r => r.Summary.Errors),
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

  var score = AUDIT.ClusterSummary.Successes / (
      AUDIT.ClusterSummary.Successes +
      AUDIT.ClusterSummary.Warnings +
      AUDIT.ClusterSummary.Errors);
  score = Math.round(score * 100);

  var clusterChart = new Chart("clusterScoreChart", {
    type: 'doughnut',
    data: {
      labels: ["Passing", "Warning", "Error"],
      datasets: [{
        data: [
          AUDIT.ClusterSummary.Successes,
          AUDIT.ClusterSummary.Warnings,
          AUDIT.ClusterSummary.Errors,
        ],
        backgroundColor: ['#8BD2DC','#f26c21','#a11f4c'],
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
          text: score + '%',
          color: '#333', //Default black
          fontStyle: 'Helvetica', //Default Arial
          sidePadding: 30 //Default 20 (as a percentage)
        }
      }
    }
  });
});

