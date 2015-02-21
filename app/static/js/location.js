$(function() {
	var placeSearch, autocomplete;

	var componentForm = {
		street_number: 'short_name',
		locality: 'long_name',
		country: 'long_name'
	};

	function initialize() {
		autocomplete = new google.maps.places.Autocomplete(
			(document.getElementById('searchLocation')),
			{types: ['geocode']});

		google.maps.event.addListener(autocomplete, 'place_changed', function() {
			fillInAddress();
		});
	}

	function fillInAddress() {
		var place = autocomplete.getPlace();
	}

	function geolocate() {
		if (navigator.geolocation) {
			navigator.geolocation.getCurrentPosition(function(position) {
				var geolocation = new google.maps.LatLng(
					position.coords.latitude, position.coords.longitude);
				autocomplete.setBounds(new google.maps.LatLngBounds(geolocation, geolocation));
			});
		}
	}

	initialize();

	$('#searchLocation').on('focus', geolocate);
});