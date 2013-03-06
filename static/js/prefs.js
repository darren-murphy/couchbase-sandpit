var cbuggPrefs = angular.module('cbuggPrefs', []);
cbuggPrefs.factory('cbuggPrefs', function($http) {

	function getDefaultPreferences() {
		return {
			search: {
				rowsPerPage: 15
			}
		};
	}

	function saveUserPreferences(prefs, success, error) {
		$http.post('/api/me/prefs/', prefs).
			success(function(res) {
				success(res);
			}).
			error(function(err) {
				error(err);
			});
	}

	return {
		getDefaultPreferences: getDefaultPreferences,
		saveUserPreferences: saveUserPreferences
	};

});

function PrefsCtrl($scope, $http, cbuggAuth, cbuggPage, cbuggPrefs, bAlert) {

	cbuggPage.setTitle("Preferences");
	$scope.auth = cbuggAuth.get();

	$scope.save = function() {
		cbuggPrefs.saveUserPreferences($scope.auth.prefs,
			function(res) {
				bAlert("Success", "Preferences Saved", "success");
			},
			function(err) {
				bAlert("Error",  err, "error");
		});
	};

	$scope.reset = function() {
		cbuggPrefs.saveUserPreferences({},
			function(res) {
				bAlert("Success", "Reset to System Defaults", "success");
				$scope.auth.prefs = cbuggPrefs.getDefaultPreferences();
			},
			function(err) {
				bAlert("Error",  err, "error");
		});
	};

}