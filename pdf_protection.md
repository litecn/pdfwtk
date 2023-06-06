PDF Permissions from tcpdf

```php
	/**
	 * Return the permission code used on encryption (P value).
	 * @param $permissions (Array) the set of permissions (specify the ones you want to block).
	 * @param $mode (int) encryption strength: 0 = RC4 40 bit; 1 = RC4 128 bit; 2 = AES 128 bit; 3 = AES 256 bit.
	 * @since 5.0.005 (2010-05-12)
	 * @author Nicola Asuni
	 * @public static
	 */
	public static function getUserPermissionCode($permissions, $mode=0) {
		$options = array(
			'owner' => 2, // bit 2 -- inverted logic: cleared by default
			'print' => 4, // bit 3
			'modify' => 8, // bit 4
			'copy' => 16, // bit 5
			'annot-forms' => 32, // bit 6
			'fill-forms' => 256, // bit 9
			'extract' => 512, // bit 10
			'assemble' => 1024,// bit 11
			'print-high' => 2048 // bit 12
			);
		$protection = 2147422012; // 32 bit: (01111111 11111111 00001111 00111100)
		foreach ($permissions as $permission) {
			if (isset($options[$permission])) {
				if (($mode > 0) OR ($options[$permission] <= 32)) {
					// set only valid permissions
					if ($options[$permission] == 2) {
						// the logic for bit 2 is inverted (cleared by default)
						$protection += $options[$permission];
					} else {
						$protection -= $options[$permission];
					}
				}
			}
		}
		return $protection;
	}
```