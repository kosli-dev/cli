function chooseDocsVersion(option) {
    let url = '';
    if (option.value == 'latest') {
      url = '/'
    } else {
      url = '/legacy_ref/' + option.value;
    }
    window.location.href = url;
  }